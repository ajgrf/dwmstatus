package main

// #cgo pkg-config: alsa
// #include <alsa/asoundlib.h>
import "C"
import (
	"errors"
	"fmt"
	"math"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const SND_CTL_TLV_DB_GAIN_MUTE = -9999999

func Volume(ch chan<- string) {
	mixer, err := NewMixer("default", "Master")
	if err != nil {
		panic(err)
	}
	defer mixer.Close()

	tc := make(chan string)
	defer close(tc)
	go Timeout(5*time.Second, tc, ch)

	for {
		mixer.Wait()
		volume, muted, err := mixer.Volume()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		} else if muted {
			tc <- "V:mute"
		} else {
			tc <- fmt.Sprintf("V:%3d%%", volume)
		}
	}
}

// Mixer represents an ALSA mixer to query for audio volume levels. A Mixer
// should be created by a call to NewMixer and closed when it's no longer
// needed.
type Mixer struct {
	handle *C.snd_mixer_t
	elem   *C.snd_mixer_elem_t
	epfd   int // epoll fd for polling for changes
}

// NewMixer returns a Mixer struct corresponding to the given ALSA card and
// mixerName.
func NewMixer(card, mixerName string) (*Mixer, error) {
	var handle *C.snd_mixer_t
	var elem *C.snd_mixer_elem_t

	// C strings need their own variable so that they can be freed
	var cardC = C.CString(card)
	defer C.free(unsafe.Pointer(cardC))
	var mixerNameC = C.CString(mixerName)
	defer C.free(unsafe.Pointer(mixerNameC))

	// simple mixer element id, needed to correctly set elem
	var sid *C.snd_mixer_selem_id_t

	// set up sid
	if C.snd_mixer_selem_id_malloc(&sid) < 0 {
		return nil, errors.New("alsa: allocating simple mixer id failed")
	}
	defer C.snd_mixer_selem_id_free(sid)
	C.snd_mixer_selem_id_set_index(sid, 0)
	C.snd_mixer_selem_id_set_name(sid, mixerNameC)

	// set up mixer handle
	if C.snd_mixer_open(&handle, 0) < 0 {
		return nil, errors.New("alsa: opening mixer failed")
	}
	if C.snd_mixer_attach(handle, cardC) < 0 {
		C.snd_mixer_close(handle)
		return nil, errors.New("alsa: attaching mixer failed")
	}
	if C.snd_mixer_selem_register(handle, nil, nil) < 0 {
		C.snd_mixer_close(handle)
		return nil, errors.New("alsa: registering mixer failed")
	}
	if C.snd_mixer_load(handle) < 0 {
		C.snd_mixer_close(handle)
		return nil, errors.New("alsa: loading mixer failed")
	}

	// set elem, which is what we need to get the volume
	if elem = C.snd_mixer_find_selem(handle, sid); elem == nil {
		C.snd_mixer_close(handle)
		return nil, errors.New("alsa: finding simple mixer element failed")
	}

	return &Mixer{handle: handle, elem: elem, epfd: -1}, nil
}

// Volume returns the Mixer's volume as a percentage, whether it's currently
// muted, and an error on failure (or nil on success).
func (m *Mixer) Volume() (volume int, muted bool, err error) {
	var value C.long = -1
	var minVol, maxVol C.long
	if C.snd_mixer_selem_get_playback_dB_range(m.elem, &minVol, &maxVol) < 0 {
		// percentage is linear if dB info is unavailable

		// get volume range
		if C.snd_mixer_selem_get_playback_volume_range(m.elem, &minVol, &maxVol) < 0 {
			err = errors.New("alsa: getting volume range failed")
			return
		}

		// get volume
		if C.snd_mixer_selem_get_playback_volume(m.elem, 0, &value) < 0 {
			err = errors.New("alsa: getting playback volume failed")
			return
		}

		// compute volume percentage
		value -= minVol
		maxVol -= minVol
		volume = int((100 * float64(value) / float64(maxVol)) + 0.5)
	} else {
		// percentage is normalized for human perception if possible

		// get dB range
		if C.snd_mixer_selem_get_playback_dB(m.elem, 0, &value) < 0 {
			err = errors.New("alsa: getting playback dB failed")
			return
		}

		if maxVol-minVol <= 24*100 {
			// small dB ranges (< 24 dB) use linear percentage anyway
			value -= minVol
			maxVol -= minVol
			volume = int((100 * float64(value) / float64(maxVol)) + 0.5)
		} else {
			// normalize volume percentage
			normalized := math.Pow(10, float64(value-maxVol)/6000)
			if minVol != SND_CTL_TLV_DB_GAIN_MUTE {
				minNorm := math.Pow(10, float64(minVol-maxVol)/6000)
				normalized = (normalized - minNorm) / (1 - minNorm)
			}
			volume = int((normalized * 100) + 0.5)
		}
	}

	// determine whether it's muted
	var mutedC C.int = -1
	if C.snd_mixer_selem_has_playback_switch(m.elem) == 1 {
		if C.snd_mixer_selem_get_playback_switch(m.elem, 0, &mutedC) < 0 {
			err = errors.New("alsa: getting playback switch failed")
			return
		}
	}
	if mutedC == 0 {
		muted = true
	}

	return
}

// Wait waits for the mixer to change volume before returning an error, if any.
func (m *Mixer) Wait() error {
	// ALSA uses a file descriptor-based interface to poll for changes

	// check for uninitialized epoll fd
	if m.epfd <= 2 {
		// get alsa file descriptors to monitor
		numFds := C.snd_mixer_poll_descriptors_count(m.handle)
		pollfds := make([]C.struct_pollfd, numFds)
		if C.snd_mixer_poll_descriptors(m.handle, &pollfds[0], C.uint(numFds)) < 0 {
			return errors.New("alsa: getting mixer poll descriptors failed")
		}

		// create epoll instance, saving epfd in m for reuse
		var err error
		m.epfd, err = syscall.EpollCreate1(0)
		if err != nil {
			return err
		}

		// add file descriptors to epoll instance for monitoring
		for i := range pollfds {
			syscall.EpollCtl(m.epfd, syscall.EPOLL_CTL_ADD,
				int(pollfds[i].fd),
				&syscall.EpollEvent{Events: syscall.EPOLLIN,
					Fd: int32(pollfds[i].fd)})
		}
	}

	// wait for changes
	events := make([]syscall.EpollEvent, 1)
wait:
	_, err := syscall.EpollWait(m.epfd, events, -1)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok && errno == syscall.EINTR {
			// wait again if the only error is an interrupt
			goto wait
		} else {
			return err
		}
	}
	if C.snd_mixer_handle_events(m.handle) < 0 {
		return errors.New("alsa: handling mixer callback events failed")
	}
	return nil
}

// Close cleans up resources used by the Mixer and returns an error, if any.
func (m *Mixer) Close() error {
	if C.snd_mixer_close(m.handle) < 0 {
		return errors.New("alsa: closing mixer handle failed")
	}
	return nil
}
