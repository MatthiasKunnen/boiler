package steamworkshop

import "bytes"

type depthTracker struct {
	TagName    []byte
	depth      int
	onComplete func()
}

func (d *depthTracker) Reset(tagName []byte, onComplete func()) {
	d.TagName = tagName
	d.depth = 1
	d.onComplete = onComplete
}

func (d *depthTracker) Increase(tagName []byte) {
	if len(d.TagName) == 0 {
		return
	}

	if !bytes.Equal(tagName, d.TagName) {
		return
	}
	d.depth++
}

// Decrease decreases the depth by 1 and will return true if the depth reaches 0.
func (d *depthTracker) Decrease(tagName []byte) {
	if len(d.TagName) == 0 {
		return
	}

	if !bytes.Equal(tagName, d.TagName) {
		return
	}
	d.depth--

	if d.depth > 0 {
		return
	}

	d.onComplete()
	d.onComplete = nil
	d.TagName = nil
}
