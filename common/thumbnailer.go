package common

import (
	"image"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"strconv"

	"github.com/Azareal/Gosora/query_gen"
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
			user, err := Users.Get(uid)
			if err != nil {
				return errors.WithStack(err)
			}

			// Has the avatar been removed or already been processed by the thumbnailer?
			if len(user.RawAvatar) < 2 || user.RawAvatar[1] == '.' {
				_, _ = acc.Delete("users_avatar_queue").Where("uid = ?").Run(uid)
				return nil
			}
			_, err = os.Stat("./uploads/avatar_" + strconv.Itoa(user.ID) + user.RawAvatar)
			if os.IsNotExist(err) {
				_, _ = acc.Delete("users_avatar_queue").Where("uid = ?").Run(uid)
				return nil
			} else if err != nil {
				return errors.WithStack(err)
			}

			// This means it's an external image, they aren't currently implemented, but this is here for when they are
			if user.RawAvatar[0] != '.' {
				return nil
			}
			/*if user.RawAvatar == ".gif" {
				return nil
			}*/
			if user.RawAvatar != ".png" && user.RawAvatar != ".jpg" && user.RawAvatar != ".jpeg" && user.RawAvatar != ".gif" {
				return nil
			}

			err = Thumbnailer.Resize(user.RawAvatar[1:], "./uploads/avatar_"+strconv.Itoa(user.ID)+user.RawAvatar, "./uploads/avatar_"+strconv.Itoa(user.ID)+"_tmp"+user.RawAvatar, "./uploads/avatar_"+strconv.Itoa(user.ID)+"_w48"+user.RawAvatar, 48)
			if err != nil {
				return errors.WithStack(err)
			}

			err = user.ChangeAvatar("." + user.RawAvatar)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = acc.Delete("users_avatar_queue").Where("uid = ?").Run(uid)
			return errors.WithStack(err)
		})
		if err != nil {
			LogError(err)
		}
		if err = acc.FirstError(); err != nil {
			LogError(err)
		}
	}
}

var Thumbnailer ThumbnailerInt

type ThumbnailerInt interface {
	Resize(format string, inPath string, tmpPath string, outPath string, width int) error
}

type RezThumbnailer struct {
}

func (thumb *RezThumbnailer) Resize(format string, inPath string, tmpPath string, outPath string, width int) error {
	// TODO: Sniff the aspect ratio of the image and calculate the dest height accordingly, bug make sure it isn't excessively high
	return nil
}

func (thumb *RezThumbnailer) resize(format string, inPath string, outPath string, width int, height int) error {
	return nil
}

// ! Note: CaireThumbnailer can't handle gifs, so we'll have to either cap their sizes or have another resizer deal with them
type CaireThumbnailer struct {
}

func NewCaireThumbnailer() *CaireThumbnailer {
	return &CaireThumbnailer{}
}

func precodeImage(format string, inPath string, tmpPath string) error {
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
	if format == "gif" {
		return gif.Encode(outFile, img, nil)
	}
	return jpeg.Encode(outFile, img, nil)
}

func (thumb *CaireThumbnailer) Resize(format string, inPath string, tmpPath string, outPath string, width int) error {
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
