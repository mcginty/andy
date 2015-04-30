package main

import (
  "github.com/nfnt/resize"
  "strings"
  "image"
  "image/png"
  "log"
  "os"
  "path/filepath"
  "github.com/spf13/cobra"
  "github.com/fatih/color"
  "fmt"
  "errors"
)

type dpi float64

type DrawableInfo struct {
  ResFolder string
  Density dpi
  Filename string
}

const (
  LDPI    = 3
  MDPI    = 4
  HDPI    = 6
  XHDPI   = 8
  XXHDPI  = 12
  XXXHDPI = 16
)

var (
  folderToDensity = map[string]dpi{
    //"drawable-ldpi":    LDPI,
    "drawable-mdpi":    MDPI,
    "drawable-hdpi":    HDPI,
    "drawable-xhdpi":   XHDPI,
    "drawable-xxhdpi":  XXHDPI,
    "drawable-xxxhdpi": XXXHDPI,
  }

  densityToFolder = map[dpi]string{
    //LDPI:   "drawable-ldpi",
    MDPI:   "drawable-mdpi",
    HDPI:   "drawable-hdpi",
    XHDPI:  "drawable-xhdpi",
    XXHDPI: "drawable-xxhdpi",
    XXXHDPI:"drawable-xxxhdpi",
  }

  densityPriorityList = []string{
    "drawable-xxxhdpi",
    "drawable-xxhdpi",
    "drawable-xhdpi",
    "drawable-hdpi",
    "drawable-mdpi",
    //"drawable-ldpi",
  }

  green = color.New(color.FgGreen).SprintfFunc()
)

func fileExists(file string) bool {
  fi, err := os.Stat(file)
  return err == nil && fi.Mode().IsRegular()
}

func dirExists(dir string) bool {
  fi, err := os.Stat(dir)
  return err == nil && fi.IsDir()
}

func pathExists(path string) bool {
  _, err := os.Stat(path)
  return err == nil
}

func guessResFolder() (folder string, err error) {
  ResDirGuesses := []string{ "res", "src/main/res" }
  for _, guess := range ResDirGuesses {
    if dirExists(guess) {
      return guess, nil
    }
  }

  return "", errors.New("no folder found")
}

func extractResFolder(path string) (folder string, err error) {
  fmt.Printf("extractResFolder(\"%s\")\n", path)
  folders := strings.Split(filepath.ToSlash(path), "/")
  for i, folder := range folders {
    if folder == "res" {
      return filepath.FromSlash(strings.Join(folders[:i+1], "/")), nil
    }
  }

  return "", errors.New("no folder found")
}

func extractDensity(path string) (density dpi, err error) {
  folders := strings.Split(filepath.ToSlash(path), "/")
  for _, folder := range folders {
    if density = folderToDensity[folder]; density > 0 {
      return
    }
  }

  err = errors.New("no density found")
  return
}

func tryGetAbsPath(path string) (absPath string) {
  absPath, err := filepath.Abs(path)
  if err != nil {
    return path
  } else {
    return absPath
  }
}

func findHighestDensity(resFolder string, filename string) (density dpi, err error) {
  for _, folder := range densityPriorityList {
    if fileExists(filepath.Join(resFolder, folder, filename)) {
      return folderToDensity[folder], nil
    }
  }
  return 0, errors.New("no density found")
}

func getDrawableInfo(path string) (info DrawableInfo, err error) {
  absPath := tryGetAbsPath(path)
  var resFolder, filename string
  var density dpi
  if fileExists(absPath) {
    resFolder, err = extractResFolder(absPath)
    if err != nil { return }
    density, err = extractDensity(absPath)
    if err != nil { return }
    _, filename = filepath.Split(path)
  } else {
    resFolder, err = guessResFolder()
    if err != nil { return }
    density, err = findHighestDensity(resFolder, path)
    if err != nil { return }
    filename = path
  }

  return DrawableInfo{ResFolder: tryGetAbsPath(resFolder), Filename: filename, Density: density}, nil
}

func getDimens(img *image.Image) (width int, height int) {
  return (*img).Bounds().Max.X - (*img).Bounds().Min.X, (*img).Bounds().Max.Y - (*img).Bounds().Min.Y
}

func resizeToFolders(drawableInfo *DrawableInfo, img *image.Image) {
  var startingDensity int
  for i, folder := range densityPriorityList {
    if (folderToDensity[folder] == (*drawableInfo).Density) {
      startingDensity = i+1
      break
    }
  }

  if startingDensity < len(densityPriorityList) {
    for _, folder := range densityPriorityList[startingDensity:] {
      resizeTo(drawableInfo, img, folder)
    }
  }
}

func resizeTo(drawableInfo *DrawableInfo, img *image.Image, folder string) {
  targetDensity := folderToDensity[folder]
  targetPath := filepath.Join((*drawableInfo).ResFolder, folder, (*drawableInfo).Filename)
  width, _ := getDimens(img)
  resized := resize.Resize(uint(float64(width)*float64(targetDensity)/float64((*drawableInfo).Density)), 0, *img, resize.Lanczos3)
  out, err := os.Create(targetPath)
  if err != nil {
    log.Fatal(err)
  }
  defer out.Close()

  png.Encode(out, resized);
  fmt.Printf("%s %s\n", green("generated"), targetPath)
}

func main() {
  var dpitizeCmd = &cobra.Command{
    Use: "dpi [asset filename]",
    Short: "Take an asset and resize it for various densities.",
    Run: func(cmd *cobra.Command, args []string) {
      if len(args) != 1 {
        log.Fatal("need a filename.")
      }
      drawableInfo, err := getDrawableInfo(args[0])
      if err != nil {
        log.Fatal(err)
      }
      //fmt.Printf("%s: %s, %s: %s\n", green("res dir"), drawableInfo.ResFolder, green("density"), drawableInfo.Density)

      //absPath, err := filepath.Abs(args[0])
      //fmt.Printf("%s %s\n", green("opening"), absPath)
      assetPath := filepath.Join(drawableInfo.ResFolder, densityToFolder[drawableInfo.Density], drawableInfo.Filename)
      fmt.Printf("%s %s\n", green("source"), assetPath)
      file, err := os.Open(assetPath)
      if err != nil { log.Fatal(err) }

      img, err := png.Decode(file)
      if err != nil { log.Fatal(err) }
      file.Close()

      resizeToFolders(&drawableInfo, &img)
    },
  }

  var rootCmd = &cobra.Command{Use: "andy"}
  rootCmd.AddCommand(dpitizeCmd)
  rootCmd.Execute()
}
