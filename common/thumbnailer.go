package common

var Thumbnailer ThumbnailerInt

type ThumbnailerInt interface {
}

type RezThumbnailer struct {
}

func (thumb *RezThumbnailer) Resize(path string, width int) error {
	// TODO: Sniff the aspect ratio of the image and calculate the dest height accordingly, bug make sure it isn't excessively high
	return nil
}

func (thumb *RezThumbnailer) resize(path string, width int, height int) error {
	return nil
}

/*
type LilliputThumbnailer struct {

}

type ResizeThumbnailer struct {

}
*/
