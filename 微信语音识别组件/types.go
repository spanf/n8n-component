package wechat_speech_recognition

const (
    FormatAMR   = "amr"
    FormatSpeex = "speex"
    FormatMP3   = "mp3"
    FormatWAV   = "wav"
)

const (
    LanguageZhCN = "zh_CN"
    LanguageEnUS = "en_US"
)

type UploadResponse struct {
    MediaID string `json:"media_id"`
}

type RecognitionResponse struct {
    Text string `json:"text"`
}
