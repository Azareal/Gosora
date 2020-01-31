package common

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strconv"

	"golang.org/x/image/tiff"

	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

func ThumbTask(thumbChan chan bool) {
	for {
		// Put this goroutine to sleep until we have work to do
		<-thumbChan

		// TODO: Use a real queue
		// TODO: Transactions? Self-repairing?
		acc := qgen.NewAcc()
		err := acc.Select("users_avatar_queue").Columns("uid").Limit("0,5").EachInt(func(uid int) error {
			// TODO: Do a bulk user fetch instead?
			u, err := Users.Get(uid)
			if err != nil {
				return errors.WithStack(err)
			}

			// Has the avatar been removed or already been processed by the thumbnailer?
			if len(u.RawAvatar) < 2 || u.RawAvatar[1] == '.' {
				_, _ = acc.Delete("users_avatar_queue").Where("uid=?").Run(uid)
				return nil
			}
			_, err = os.Stat("./uploads/avatar_" + strconv.Itoa(u.ID) + u.RawAvatar)
			if os.IsNotExist(err) {
				_, _ = acc.Delete("users_avatar_queue").Where("uid=?").Run(uid)
				return nil
			} else if err != nil {
				return errors.WithStack(err)
			}

			// This means it's an external image, they aren't currently implemented, but this is here for when they are
			if u.RawAvatar[0] != '.' {
				return nil
			}
			/*if user.RawAvatar == ".gif" {
				return nil
			}*/
			if u.RawAvatar != ".png" && u.RawAvatar != ".jpg" && u.RawAvatar != ".jpe" && u.RawAvatar != ".jpeg" && u.RawAvatar != ".jif" && u.RawAvatar != ".jfi" && u.RawAvatar != ".jfif" && u.RawAvatar != ".gif" && u.RawAvatar != "tiff" && u.RawAvatar != "tif" {
				return nil
			}

			ap := "./uploads/avatar_"
			err = Thumbnailer.Resize(u.RawAvatar[1:], ap+strconv.Itoa(u.ID)+u.RawAvatar, ap+strconv.Itoa(u.ID)+"_tmp"+u.RawAvatar, ap+strconv.Itoa(u.ID)+"_w48"+u.RawAvatar, 48)
			if err != nil {
				return errors.WithStack(err)
			}

			err = u.ChangeAvatar("." + u.RawAvatar)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = acc.Delete("users_avatar_queue").Where("uid=?").Run(uid)
			return errors.WithStack(err)
		})
		if err != nil {
			LogError(err)
		}

		/*
			err := acc.Select("attach_image_queue").Columns("attachID").Limit("0,5").EachInt(func(attachID int) error {
				return nil

				_, err = acc.Delete("attach_image_queue").Where("attachID = ?").Run(uid)
			}
		*/
		if err = acc.FirstError(); err != nil {
			LogError(err)
		}
	}
}

var Thumbnailer ThumbnailerInt

type ThumbnailerInt interface {
	Resize(format, inPath, tmpPath, outPath string, width int) error
}

type RezThumbnailer struct {
}

func (thumb *RezThumbnailer) Resize(format, inPath, tmpPath, outPath string, width int) error {
	// TODO: Sniff the aspect ratio of the image and calculate the dest height accordingly, bug make sure it isn't excessively high
	return nil
}

func (thumb *RezThumbnailer) resize(format, inPath, outPath string, width, height int) error {
	return nil
}

// ! Note: CaireThumbnailer can't handle gifs, so we'll have to either cap their sizes or have another resizer deal with them
type CaireThumbnailer struct {
}

func NewCaireThumbnailer() *CaireThumbnailer {
	return &CaireThumbnailer{}
}

func precodeImage(format, inPath, tmpPath string) error {
	imageFile, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer imageFile.Close()

	img, _, err := image.Decode(imageFile)
	if err != nil {
		return err
	}

	outFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// TODO: Make sure animated gifs work after being encoded
	switch format {
	case "gif":
		return gif.Encode(outFile, img, nil)
	case "png":
		return png.Encode(outFile, img)
	case "tiff", "tif":
		return tiff.Encode(outFile, img, nil)
	}
	return jpeg.Encode(outFile, img, nil)
}

func (thumb *CaireThumbnailer) Resize(format, inPath, tmpPath, outPath string, width int) error {
	err := precodeImage(format, inPath, tmpPath)
	if err != nil {
		return err
	}
	return nil

	// TODO: Caire doesn't work. Try something else. Or get them to fix the index out of range. We get enough wins from re-encoding as jpeg anyway
	/*imageFile, err := os.Open(tmpPath)
	if err != nil {
		return err
	}
	defer imageFile.Close()

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	p := &caire.Processor{NewWidth: width, Scale: true}
	return p.Process(imageFile, outFile)*/
}

/*
type LilliputThumbnailer struct {

}
*/
