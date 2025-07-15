# Youtube Media Downloader
This is a GUI application which helps users download videos (and / or) audios from youtube including shorts,
What makes this app special is that it allows you to select from multiple codecs for videos and audios,
Examples: AV1, VP9, H.264, OPUS, Dolby Digial Plus, AAC, ..
you can also select different video resolutions and
choose to merge video and audio or not.

It uses yt-dlp to detect codecs and their corresponding resolutions,
and download them.
The best part is that the app's gui is built purely in GO making it performant.

![App Screenshot](screenshot.png)

# Build From Source
- Arch Linux:
  ```bash
  sudo pacman -S git base-devel libxrandr libxi libxcursor libxinerama go yt-dlp ffmpeg --noconfirm --needed
  cd ~/Downloads
  git clone https://github.com/TejasPersonal/youtube-media-downloader
  cd youtube-media-downloader
  go mod tidy
  go build -o bin/

  # run
  bin/youtube-media-downloader
  ```
