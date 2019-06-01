package main

import (
	"image"
	_ "image/png"
	"net/http"
	botApi "tgApi/bot"
)

func main() {
	bot := botApi.NewBot("611561011:AAGmEzv8GPOERZcru038nmFxqC8O_geiC4o")

	img_f, _ := http.Get("https://cdn-images-1.medium.com/max/1600/1*Pp5_-7MlwmssrseF50AhMQ.png")

	img, _, _ := image.Decode(img_f.Body)

	bot.SendPhoto(136313212, img)

	bot.Start()
}
