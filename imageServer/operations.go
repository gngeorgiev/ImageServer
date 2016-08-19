package imageServer

import "gopkg.in/h2non/bimg.v1"

type ResizeResult struct {
	Image Image
	Error error
}

func resize(buffer []byte, params ImageParams, doneChan chan ResizeResult) {
	workerId := <-workers
	res := &ResizeResult{}

	opts := params.toBimgOptions()
	b, err := bimg.Resize(buffer, opts)
	if err != nil {
		res.Error = err
	}

	newImg, err := NewImage(b)
	if err != nil {
		res.Error = err
	}

	res.Image = newImg

	workers <- workerId
	doneChan <- *res
}
