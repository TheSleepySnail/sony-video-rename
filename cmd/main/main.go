package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/sirupsen/logrus"
)

func check(e error) {
	if e != nil {
		fmt.Print(e.Error())
		os.Exit(1)
	}
}

type Config struct {
	dryRun        bool
	ignoreMissing bool
	camera        bool
	original      bool
}

type ToStringFormatter struct{}

func (f *ToStringFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer = &bytes.Buffer{}
	b.WriteString(entry.Message)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func main() {
	var version = "1.0.0-SNAPSHOT"
	var example = "Example usage: SonyVideoRename -d -s=MySuffix -o=false -f ~/MyVideos"
	var helpFlag = flag.Bool("h", false, "Show this help")
	var folderFlag = flag.String("f", ".", "Path to the folder where the files are. If not set -> Same folder where exe is executed")
	var suffixFlag = flag.String("s", "", "Optional suffix to be added to the file name")
	var timeFlag = flag.String("t", "+0h", "Optional time correction. E.g. -t=+0h1m2s")
	var dryRunFlag = flag.Bool("d", false, "Dry run, just print out what this tool here would do without actually renaming files")
	var ignoreMissingFlag = flag.Bool("i", false, "Ignore missing files. By default if an MP4 was not found I will not do anything.")
	var originalFlag = flag.Bool("o", true, "Adding original file name to the new file name. E.g. _(C0001)")
	var cameraFlag = flag.Bool("c", true, "Adding camera name to the new file name. E.g. _(XDR-200).")
	var debugFlag = flag.Bool("v", false, "More logging")
	flag.Parse()

	if *helpFlag {
		fmt.Println("Version " + version)
		flag.PrintDefaults()
		fmt.Println(example)
		os.Exit(0)
	}

	if *folderFlag == "" {
		fmt.Println("Version " + version)
		fmt.Println("Missing mandatory arguments")
		fmt.Println(example)
		flag.PrintDefaults()
		os.Exit(1)
	}

	log := &logrus.Logger{
		Out:       os.Stderr,
		Level:     logrus.InfoLevel,
		Formatter: &ToStringFormatter{},
	}

	log.Info("Version " + version)
	if *debugFlag {
		log.Level = logrus.DebugLevel
	} else {
		log.Level = logrus.InfoLevel
	}

	config := Config{dryRun: *dryRunFlag, ignoreMissing: *ignoreMissingFlag, camera: *cameraFlag, original: *originalFlag}

	folder, _ := filepath.Abs(*folderFlag)
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if len(files) == 0 {
		log.Error("No files found in " + folder)
		os.Exit(1)
	} else {
		log.Infof("Found %v files", len(files))
	}

	log.Info("Configuration:")
	log.Infof(" * Folder: %s", folder)

	if config.dryRun {
		log.Info(" * Doing a dry run")
	}
	if config.original {
		log.Info(" * Adding original file name")
	}
	if config.camera {
		log.Info(" * Adding camera name")
	}

	reader := bufio.NewReader(os.Stdin)
	suffix := *suffixFlag

	if suffix != "" {
		log.Info(" * Adding suffix " + suffix + " to filename")
	}

	timeShift := *timeFlag
	timePlusMinus := timeShift[0:1]
	if timePlusMinus != "+" && timePlusMinus != "-" {
		log.Errorf("Unknown time shift command '%s'. Only '+' or '-' allowed", timePlusMinus)
		os.Exit(1)

	}
	timeShiftDuration, err := time.ParseDuration(timeShift[1:])
	check(err)

	log.Infof(" * Timeshift=%s%vs (%s)", timePlusMinus, timeShiftDuration.Seconds(), timeShift)

	log.Info("Do you want to continue?")
	doContinue, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(doContinue)) == "y" {
	} else {
		os.Exit(-1)
	}

	var renamedFiles = 0
	var missingFiles = 0
	for _, file := range files {
		var fileName = filepath.Join(folder, file.Name())
		if strings.HasSuffix(fileName, ".XML") {
			dir := filepath.Dir(fileName)
			cleanOriginalFileName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
			cleanOriginalFileName = strings.Replace(cleanOriginalFileName, "M01", "", 1)
			log.Debug("Parsing file: " + fileName + ". " + cleanOriginalFileName)
			dat, err := ioutil.ReadFile(fileName)
			check(err)

			var xml = string(dat)
			doc, err := xmlquery.Parse(strings.NewReader(xml))
			creationDataString := xmlquery.Find(doc, "//CreationDate")[0].SelectAttr("value")

			log.Debug("Found CreationDate: " + creationDataString + ". Trying to parse it")
			layout := "2006-01-02T15:04:05Z07:00"

			t, _ := time.Parse(layout, creationDataString)

			var cameraString = ""
			if config.camera {
				cameraString = xmlquery.Find(doc, "//Device")[0].SelectAttr("modelName")
			}

			if timeShiftDuration.Seconds() == 0 {
				// Doing nothing
			} else {
				oldTime := t
				if timePlusMinus == "+" {
					t = t.Add(timeShiftDuration)
				} else {
					t = t.Add(-timeShiftDuration)
				}
				log.Infof("%s: OldTime=%v, NewTime=%v", cleanOriginalFileName, oldTime.Format("15h04m05s"), t.Format("15h04m05s"))
			}

			prefix := t.Format("20060102_150405")
			prefix = prefix + " - "
			newFileNamePrefix := prefix
			if config.original {
				newFileNamePrefix = newFileNamePrefix + "(" + cleanOriginalFileName + ")"
			}
			if config.camera {
				newFileNamePrefix = newFileNamePrefix + "(" + cameraString + ")"
			}
			if suffix != "" {
				newFileNamePrefix = newFileNamePrefix + suffix
			}

			log.Debug(newFileNamePrefix)

			oldFileXml := fileName
			oldFileMp4 := dir + "/" + cleanOriginalFileName + ".MP4"
			newFileXml := dir + "/" + newFileNamePrefix + ".xml"
			newFileMp4 := dir + "/" + newFileNamePrefix + ".mp4"
			if _, err := os.Stat(oldFileMp4); err == nil {
				// // path/to/whatever exists
				log.Info(cleanOriginalFileName + ": Renaming\n    " + oldFileMp4 + " to\n    " + newFileMp4)
				if !config.dryRun {
					os.Rename(oldFileMp4, newFileMp4)
					renamedFiles++
				}
				log.Info(cleanOriginalFileName + ": Renaming\n    " + oldFileXml + " to\n    " + newFileXml)
				if !config.dryRun {
					os.Rename(oldFileXml, newFileXml)
					renamedFiles++
				}
			} else {
				log.Error("File " + oldFileMp4 + " not found: " + err.Error())
				if !config.ignoreMissing {
					os.Exit(1)
				}
				missingFiles++
			}
		}
	}
	if missingFiles > 0 {
		log.Infof("I renamed %v files but failed to rename %v missing files.", renamedFiles, missingFiles)
	} else {
		log.Infof("I renamed %v files.", renamedFiles)
	}
	_, _ = reader.ReadString('\n')
}
