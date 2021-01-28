package jobs

import (
	"bytes"

	"github.com/signintech/gopdf"
)

func imageToPDF(imageBytes []byte) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	image, err := gopdf.ImageHolderByBytes(imageBytes)
	if err != nil {
		return nil, err
	}

	err = pdf.ImageByHolder(image, 0, 0, nil)
	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)
	pdf.Write(buffer)
	return buffer.Bytes(), nil
}
