# Video rename tool for sony cameras

When recording videos, most Sony cameras create a sidecar file for each video containing information about the creation date, camera model, etc...
This tool helps with renaming of the video files to include the timestamps in the file name.
E.g. `C0001.MP4` will be renamed to `20210914_095133 - (C0001)(ILCE-7M3)` 
An optional additional suffix can be added to the end of the filename.
**Tested with ILCE-7M3, FDR-X3000**

## Remark
It seems like after powering on a Sony camera on, it is always looking for the next free number as the next filename.
If you have recorded C0001, C0002, C0003 and you delete the C0002 -> the next time the camera is powered on it will use C0002 for the next video being recorded. Renaming the files using the creation information from the sidecar file will help sorting the videos.

## Usage

> **Warning**
> This tool is renaming the original files. Make sure you have a backup!

Easiest way is to copy the executable into the folder where the videos files are located and run it from there in a terminal / command line

```
Version 1.0.0
  -c    Adding camera name to the new file name. E.g. _(XDR-200). (default true)
  -d    Dry run, just print out what this tool here would do without actually renaming files
  -f string
        Path to the folder where the files are. If not set -> Same folder where exe is executed (default ".")
  -h    Show this help
  -i    Ignore missing files. By default if an MP4 was not found I will not do anything.
  -o    Adding original file name to the new file name. E.g. _(C0001) (default true)
  -s string
        Optional suffix to be added to the file name
  -t string
        Optional time correction. E.g. -t=+0h1m2s (default "+0h")
  -v    More logging
Example usage: SonyVideoRename -d -s=MySuffix -o=false -f ~/MyVideos
```


### Examples

**Dry run**

For testing use the -d option. This will just print out what the tool would do to your files

```
SonyVideoRename -d
```


**Rename with time correction**

Adding 1 hour and 5 minutes to the time. For an offset of more then 1 day just multiply the days with 24h ;)

```
SonyVideoRename -t=+1h5m0s
```

