package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

func getCharByAlies(alies string) string {
	var char = alies
	switch (alies) {
	case "Period":
		char = "."
		break
	case "Question":
		char = "?"
		break
	case "Backslash":
		char = "\\"
		break
	case "Slash":
		char = "/"
		break
	case "Colon":
		char = ":"
		break
	case "Star":
		char = "*"
		break
	case "Quotation":
		char = "\":"
		break
	case "Head":
		char = "<"
		break
	case "Trail":
		char = ">"
		break
	case "Surprise":
		char = "!"
		break
	case "Gang":
		char = "|"
		break
	default:
		char = alies
	}
	return char
}

func getFiles(dir string) []string {
	files := make([]string, 0)

	filepath.Walk(dir, func(v string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		if path.Ext(f.Name()) == ".png" {
			files = append(files, path.Join(dir, f.Name()))
		}

		return nil
	})

	sort.Strings(files)

	return files
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
	//return "./",nil
}

func W(f *os.File, format string, s ...interface{}) {
	ss := fmt.Sprintf(format, s...)
	// fmt.Println(ss)
	f.WriteString(ss + "\n")
}

type ImagePair struct {
	Name     rune
	FileName string
	Image    image.Image
	ContentX int
	ContentY int
	ContentW int
	ContentH int
	PasteX   int
	PasteY   int
}

var gaps = flag.String("g", "10", "gap")
var dir = flag.String("d", ".", "dir")
var name = flag.String("n", "", "name")
var skip = flag.String("s", "", "remove")

var lineCharNum = 10

func NameDetect(imageName string) rune {

	ext := filepath.Ext(imageName)

	imageName = imageName[:len(imageName)-len(ext)]

	imageName = getCharByAlies(imageName)

	if *skip != "" {
		imageName = strings.Replace(imageName, *skip, "", -1)
	}

	return []rune(imageName)[0]
}

//http://www.angelcode.com/products/bmfont/doc/file_format.html
func main() {
	path,_ := GetCurrentPath()

	err4 := os.RemoveAll(path+"/ret");
	if err4 != nil {
		log.Fatal(err4);
	}


	flag.Parse()
	var gap,_ = strconv.Atoi(*gaps)

	absPath, err := filepath.Abs(*dir)
	if err != nil {
		panic(err)
	}

	if *name == "" {
		*name = filepath.Base(absPath)
	}

	files := getFiles(*dir)

	var err1 = os.Mkdir(path+"/ret",os.ModePerm)
	if err1!= nil{
		fmt.Println(err1)
	}

	images := make([]ImagePair, len(files))
	totalW := 0
	totalH := 0
	maxH := 0

	//获取最大宽和最大高
	maaxW := 0
	maaxH := 0
	for _, name := range files {
		img, err := imaging.Open(name)

		if err != nil {
			fmt.Println("图片" + name)
			panic(err)
		}

		ww := img.Bounds().Max.X
		ww += ww/gap
		hh := img.Bounds().Max.Y
		if ww > maaxW {
			maaxW = ww
		}
		if hh > maaxH {
			maaxH = hh
		}
	}

	lineMaxW := maaxW*lineCharNum
	//设置图片位置
	curW := 0
	curH := 0
	for i, name := range files {
		k := NameDetect(filepath.Base(name))

		img, err := imaging.Open(name)

		if err != nil {
			fmt.Println("图片" + name)
			panic(err)
		}

		w := img.Bounds().Max.X
		h := img.Bounds().Max.Y

		//加上字间距
		if curW + w + w/gap > lineMaxW {
			curW = 0
			curH += maaxH
		}

		CH := h
		CW := w + w/gap
		PX := curW + w/gap/2
		PY := curH
		ip := ImagePair{
			Name:     k,
			Image:    img,
			FileName: name,
			ContentX: curW,
			ContentY: curH,
			ContentH: CH,
			ContentW: CW,
			PasteX: PX,
			PasteY: PY,
		}
		curW += CW


		totalW += CW
		totalH += h
		if h > maxH {
			maxH = h
		}

		images[i] = ip
	}

	avgW := totalW / len(files)
	avgH := totalH / len(files)

	lineNum := int(len(files)/lineCharNum)+1

	dest := imaging.New(lineMaxW, lineNum*maaxH, color.Alpha{0})
	f, _ := os.Create(path+"/ret/"+*name + ".fnt")
	//
	W(f, "info face=\"Arial\" size=%d bold=0 italic=0 charset=\"\" unicode=1 stretchH=100 smooth=1 aa=1 padding=0,0,0,0 spacing=1,1 outline=0", avgH)
	W(f, "common lineHeight=%d base=%d scaleW=%d scaleH=%d pages=1 packed=0 alphaChnl=1 redChnl=0 greenChnl=0 blueChnl=0", maaxH, maaxH, totalW, maxH)
	W(f, "page id=0 file=\"%s.png\"", *name)
	W(f, "chars count=%d", len(files)+1)

	spaceSize := math.MaxInt64
	for _, pair := range images {
		img := pair.Image
		k := pair.Name
		w := pair.ContentW
		h := pair.ContentH
		px := pair.PasteX
		py := pair.PasteY
		x := pair.ContentX
		y := pair.ContentY

		dest = imaging.Paste(dest, img, image.Pt(px, py))

		fmt.Println(fmt.Sprintf("%s => %s => %d", pair.FileName, string(pair.Name), int(pair.Name)))
		if spaceSize > w+(avgW-w)/2 {
			spaceSize = w+(avgW-w)/2
		}
		W(f, "char id=%d x=%d y=%d width=%d height=%d xoffset=%d yoffset=%d xadvance=%d page=0  chnl=15 char=\"%c\"", int(k), x, y, w, h, 0, maxH-h, w, k)
	}
	W(f, "char id=%d x=%d y=%d width=%d height=%d xoffset=%d yoffset=%d xadvance=%d page=0  chnl=15 char=\"%s\"", int(' '), 0, 0, 0, spaceSize, 0, 0, spaceSize, "space")

	if err = imaging.Save(dest, path+"/ret/"+*name+".png"); err != nil {
		panic(err)
	}
}
