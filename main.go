package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Format struct {
	FPS            float32 `json:"fps"`
	Vcodec         string  `json:"vcodec"`
	Acodec         string  `json:"acodec"`
	Protocol       string  `json:"protocol"`
	URL            string  `json:"url"`
	Vbr            float32 `json:"vbr"`
	Abr            float32 `json:"abr"`
	Tbr            float32 `json:"tbr"`
	Asr            uint    `json:"asr"`
	FileSize       uint    `json:"filesize"`
	FileSizeApprox uint    `json:"filesize_approx"`
	Width          uint    `json:"width"`
	Height         uint    `json:"height"`
	Extension      string  `json:"ext"`
	DynamicRange   string  `json:"dynamic_range"`
}

type Video struct {
	Formats    []Format `json:"formats"`
	ID         string   `json:"id"`
	FullTitle  string   `json:"fulltitle"`
	WebPageURL string   `json:"webpage_url"`
}

func contains(slice []uint, value uint) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func getVideoInfo(video Video) (map[string]string, map[string]string, map[string][]uint, map[string][]uint) {
	video_codecs := map[string]string{}
	audio_codecs := map[string]string{}

	codec_widths := map[string][]uint{}
	codec_heights := map[string][]uint{}

	for _, format := range video.Formats {
		has_audio := format.Acodec != "none" && format.Acodec != ""
		has_video := format.Vcodec != "none" && format.Vcodec != ""

		if has_video {
			if !has_audio {
				codec_name := strings.SplitN(format.Vcodec, ".", 2)[0]
				if codec_name == "av01" {
					_, exists := video_codecs["AV1"]
					if !exists {
						video_codecs["AV1"] = "vcodec*=av01"
					}
					if !contains(codec_heights["AV1"], format.Height) {
						codec_heights["AV1"] = append(codec_heights["AV1"], format.Height)
						codec_widths["AV1"] = append(codec_widths["AV1"], format.Width)
					}
				} else if codec_name == "avc1" {
					_, exists := video_codecs["H.264"]
					if !exists {
						video_codecs["H.264"] = "vcodec*=avc1"
					}
					if !contains(codec_heights["H.264"], format.Height) {
						codec_heights["H.264"] = append(codec_heights["H.264"], format.Height)
						codec_widths["H.264"] = append(codec_widths["H.264"], format.Width)
					}
				} else if strings.HasPrefix(codec_name, "vp") {
					codec_version := string(codec_name[len(codec_name)-1])
					codec_name := "vp" + codec_version

					codec_name_upp := strings.ToUpper(codec_name)

					_, exists := video_codecs[codec_name_upp]
					if !exists {
						video_codecs[codec_name_upp] = fmt.Sprintf("vcodec~='^vp(0%s|%s)'", codec_version, codec_version)
						codec_heights[codec_name_upp] = []uint{format.Height}
					}
					if !contains(codec_heights[codec_name_upp], format.Height) {
						codec_heights[codec_name_upp] = append(codec_heights[codec_name_upp], format.Height)
						codec_widths[codec_name_upp] = append(codec_widths[codec_name_upp], format.Width)
					}
				} else {
					codec_name_upp := strings.ToUpper(codec_name)
					_, exists := video_codecs[codec_name_upp]
					if !exists {
						video_codecs[codec_name_upp] = "vcodec*=" + codec_name
						codec_heights[codec_name_upp] = []uint{format.Height}
					}
					if !contains(codec_heights[codec_name_upp], format.Height) {
						codec_heights[codec_name_upp] = append(codec_heights[codec_name_upp], format.Height)
						codec_widths[codec_name_upp] = append(codec_widths[codec_name_upp], format.Width)
					}
				}
			}
		} else if has_audio {
			codec_name := strings.SplitN(format.Acodec, ".", 2)[0]
			if codec_name == "mp4a" {
				_, exists := audio_codecs["AAC"]
				if !exists {
					audio_codecs["AAC"] = "acodec*=mp4a"
				}
			} else if codec_name == "ac-3" {
				_, exists := audio_codecs["Dolby Digital"]
				if !exists {
					audio_codecs["Dolby Digital"] = "acodec*=ac-3"
				}
			} else if codec_name == "ec-3" {
				_, exists := audio_codecs["Dolby Digital Plus"]
				if !exists {
					audio_codecs["Dolby Digital Plus"] = "acodec*=ec-3"
				}
			} else {
				codec_name_upp := strings.ToUpper(codec_name)
				_, exists := audio_codecs[codec_name_upp]
				if !exists {
					audio_codecs[codec_name_upp] = "acodec*=" + codec_name
				}
			}
		}
	}

	return video_codecs, audio_codecs, codec_widths, codec_heights
}

func loadImageFromURL(url string) fyne.Resource {
	resp, _ := http.Get(url)
	data, _ := io.ReadAll(resp.Body)
	return fyne.NewStaticResource("image", data)
}

func calcRes(option string, codec_heights map[string][]uint) []string {
	arr := []string{}
	if option == "BEST" {
		_, av1 := codec_heights["AV1"]
		if av1 {
			for _, height := range codec_heights["AV1"] {
				arr = append(arr, fmt.Sprintf("%d", height))
			}
			arr = append(arr, "BEST")
			return arr
		}
		_, vp9 := codec_heights["VP9"]
		if vp9 {
			for _, height := range codec_heights["VP9"] {
				arr = append(arr, fmt.Sprintf("%d", height))
			}
			arr = append(arr, "BEST")
			return arr
		}
		_, avc1 := codec_heights["H.264"]
		if avc1 {
			for _, height := range codec_heights["H.264"] {
				arr = append(arr, fmt.Sprintf("%d", height))
			}
			arr = append(arr, "BEST")
			return arr
		}
		return arr
	}
	_, codec := codec_heights[option]
	if codec {
		for _, height := range codec_heights[option] {
			arr = append(arr, fmt.Sprintf("%d", height))
		}
		arr = append(arr, "BEST")
		return arr
	}
	return arr
}

func bestVideoCodec(vcodecs map[string]string) string {
	_, av1 := vcodecs["AV1"]
	if av1 {
		return "AV1"
	}
	_, vp9 := vcodecs["VP9"]
	if vp9 {
		return "VP9"
	}
	_, avc1 := vcodecs["H.264"]
	if avc1 {
		return "H.264"
	}
	return ""
}

func bestAudioCodec(acodecs map[string]string) string {
	_, ddp := acodecs["Dolby Digital Plus"]
	if ddp {
		return "Dolby Digital Plus"
	}
	_, dd := acodecs["Dolby Digital"]
	if dd {
		return "Dolby Digital"
	}
	_, opus := acodecs["OPUS"]
	if opus {
		return "OPUS"
	}
	_, aac := acodecs["AAC"]
	if aac {
		return "AAC"
	}
	return ""
}

func isInstalled(binaryName string) bool {
	_, err := exec.LookPath(binaryName)
	return err == nil
}

func main() {
	var yt_dlp_path string
	if isInstalled("yt-dlp") {
		yt_dlp_path = "yt-dlp"
	} else {
		binary_path, _ := os.Executable()
		binary_dir := filepath.Dir(binary_path)
		yt_dlp_path = binary_dir + string(filepath.Separator) + "yt-dlp"
	}

	videos_path := GetVideoFolder()
	var video Video

	var video_codecs map[string]string
	// var codec_widths map[string][]uint
	var codec_heights map[string][]uint

	var audio_codecs map[string]string

	ytApp := app.New()
	window := ytApp.NewWindow("Youtube Media Downloader")
	url_input := widget.NewEntry()
	url_input.SetPlaceHolder("URL")
	// url_input.Text = "https://youtu.be/rJNBGqiBI7s?si=quoUqCGSJaUDFywQ"

	thumbnail := canvas.NewImageFromResource(nil)
	thumbnail.FillMode = canvas.ImageFillContain
	thumbnail.SetMinSize(fyne.NewSize(0, 220))

	title := widget.NewLabel("")
	title.Wrapping = fyne.TextWrapBreak

	loading := widget.NewLabel("Loading..")
	loading.Wrapping = fyne.TextWrapBreak
	loading.Hide()

	progress := widget.NewLabel("")
	progress.Wrapping = fyne.TextWrapBreak
	progress.Hide()

	audio_progress := widget.NewLabel("")
	audio_progress.Wrapping = fyne.TextWrapBreak
	audio_progress.Hide()

	preview_box := container.NewVBox(thumbnail, title)

	resolutions := widget.NewSelect(nil, nil)

	video_codec := widget.NewSelect(nil, func(option string) {
		resolutions.Options = calcRes(option, codec_heights)
		resolutions.SetSelectedIndex(len(resolutions.Options) - 1)
	})
	var vcodec_arr []string

	audio_codec := widget.NewSelect(nil, nil)
	var acodec_arr []string

	video_option_label := widget.NewLabel("Video Settings: ")
	video_option_label.Wrapping = fyne.TextWrapBreak
	video_codec_label := widget.NewLabel("Codec: ")
	video_codec_label.Wrapping = fyne.TextWrapBreak
	video_res_label := widget.NewLabel("Resolution: ")
	video_res_label.Wrapping = fyne.TextWrapBreak

	video_options := container.NewVBox(video_option_label, video_codec_label, video_codec, video_res_label, resolutions)

	audio_option_label := widget.NewLabel("Audio Settings: ")
	audio_option_label.Wrapping = fyne.TextWrapBreak
	audio_codec_label := widget.NewLabel("Codec: ")
	audio_codec_label.Wrapping = fyne.TextWrapBreak

	audio_options := container.NewVBox(audio_option_label, audio_codec_label, audio_codec)

	options := container.NewGridWithColumns(2, video_options, audio_options)

	merge := widget.NewCheck("Merge video and audio", nil)
	merge.SetChecked(true)

	directory_label := widget.NewLabel("Download folder:")
	directory_label.Wrapping = fyne.TextWrapBreak

	directory := widget.NewEntry()
	directory.SetText(videos_path)

	download_button := widget.NewButton("Download", nil)
	download_button.OnTapped = func() {
		download_button.Disable()
		var selected_video_codec string
		if video_codec.Selected == "BEST" {
			selected_video_codec = bestVideoCodec(video_codecs)
		} else {
			selected_video_codec = video_codec.Selected
		}
		var selected_audio_codec string
		if audio_codec.Selected == "BEST" {
			selected_audio_codec = bestAudioCodec(audio_codecs)
		} else {
			selected_audio_codec = audio_codec.Selected
		}
		var selected_resolution string
		if resolutions.Selected == "BEST" {
			selected_resolution = resolutions.Options[resolutions.SelectedIndex()-1]
		} else {
			selected_resolution = resolutions.Selected
		}

		if !strings.HasSuffix(directory.Text, string(filepath.Separator)) {
			directory.Text += string(filepath.Separator)
		}

		progress.Show()

		if merge.Checked {
			settings := fmt.Sprintf(
				"bestvideo[%s][height=%s]+bestaudio[%s]",
				video_codecs[selected_video_codec],
				selected_resolution,
				audio_codecs[selected_audio_codec],
			)

			command := exec.Command(
				yt_dlp_path,
				"-f", settings,
				video.WebPageURL,
				"-o", directory.Text+"%(title)s.%(ext)s",
				"--newline",
			)
			// fmt.Println(command.Args)

			removeTerminal(command)

			go func() {
				stdout, _ := command.StdoutPipe()
				command.Start()
				scanner := bufio.NewScanner(stdout)

				for scanner.Scan() {
					line := scanner.Text()
					fyne.Do(func() {
						progress.SetText(line)
					})
				}

				fyne.Do(func() {
					progress.SetText("Download Finished!")
					download_button.Enable()
				})
			}()
		} else {
			audio_progress.Show()

			video_settings := fmt.Sprintf(
				"bestvideo[%s][height=%s]",
				video_codecs[selected_video_codec],
				selected_resolution,
			)
			audio_settings := fmt.Sprintf(
				"bestaudio[%s]",
				audio_codecs[selected_audio_codec],
			)

			video_command := exec.Command(
				yt_dlp_path,
				"-f", video_settings,
				video.WebPageURL,
				"-o", directory.Text+"%(title)s video.%(ext)s",
				"--newline",
			)
			// fmt.Println(video_command.Args)

			removeTerminal(video_command)

			go func() {
				stdout, _ := video_command.StdoutPipe()
				video_command.Start()
				scanner := bufio.NewScanner(stdout)

				for scanner.Scan() {
					line := scanner.Text()
					fyne.Do(func() {
						progress.SetText(line)
					})
				}

				fyne.Do(func() {
					progress.SetText("Download Finished!")
					download_button.Enable()
				})
			}()

			audio_command := exec.Command(
				yt_dlp_path,
				"-f", audio_settings,
				video.WebPageURL,
				"-o", directory.Text+"%(title)s audio.%(ext)s",
				"--newline",
			)
			// fmt.Println(audio_command.Args)

			removeTerminal(audio_command)

			go func() {
				stdout, _ := audio_command.StdoutPipe()
				audio_command.Start()
				scanner := bufio.NewScanner(stdout)

				for scanner.Scan() {
					line := scanner.Text()
					fyne.Do(func() {
						audio_progress.SetText(line)
					})
				}

				fyne.Do(func() {
					audio_progress.SetText("Download Finished!")
				})
			}()
		}
	}

	download_box := container.NewGridWithColumns(2, merge, download_button)
	options_box := container.NewVBox(options, directory_label, directory, download_box)

	video_box := container.NewGridWithColumns(2, preview_box, options_box)
	video_box.Hide()

	get_video_button := widget.NewButton("Get Video", nil)

	url_input.OnSubmitted = func(video_url string) {
		progress.SetText("")
		progress.Hide()
		audio_progress.SetText("")
		audio_progress.Hide()
		get_video_button.Disable()
		url_input.Disable()
		go func() {
			fyne.Do(func() {
				loading.Show()
			})

			command := exec.Command(
				yt_dlp_path,
				// "--extractor-args", "youtube:player-client=android_vr",
				"-j", video_url,
			)

			removeTerminal(command)

			json_string, err := command.Output()
			if err != nil {
				fyne.Do(func() {
					loading.Hide()
					get_video_button.Enable()
					url_input.Enable()
				})
				return
			}

			video = Video{}
			json.Unmarshal(json_string, &video)
			thumbnail_url := "https://img.youtube.com/vi/" + video.ID + "/maxresdefault.jpg"

			video_codecs, audio_codecs, _, codec_heights = getVideoInfo(video)

			vcodec_arr = []string{}
			for codec := range video_codecs {
				vcodec_arr = append(vcodec_arr, codec)
			}
			vcodec_arr = append(vcodec_arr, "BEST")
			video_codec.Options = vcodec_arr
			fyne.Do(func() {
				video_codec.SetSelectedIndex(len(video_codec.Options) - 1)
			})

			acodec_arr = []string{}
			for codec := range audio_codecs {
				acodec_arr = append(acodec_arr, codec)
			}
			acodec_arr = append(acodec_arr, "BEST")
			audio_codec.Options = acodec_arr
			fyne.Do(func() {
				audio_codec.SetSelectedIndex(len(audio_codec.Options) - 1)
			})

			thumbnail.Resource = loadImageFromURL(thumbnail_url)

			fyne.Do(func() {
				loading.Hide()
				title.SetText(video.FullTitle)
				thumbnail.Refresh()
				video_box.Show()
				url_input.Enable()
				get_video_button.Enable()
			})
		}()
	}

	get_video_button.OnTapped = func() {
		url_input.OnSubmitted(url_input.Text)
	}

	input_box := container.NewVBox(url_input)
	get_video_button_box := container.NewVBox(get_video_button)

	input_space := container.NewGridWithColumns(2, input_box, get_video_button_box)

	content := container.NewVBox(
		input_space,
		video_box,
		loading,
		progress,
		audio_progress,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(800, 500))
	window.ShowAndRun()
}
