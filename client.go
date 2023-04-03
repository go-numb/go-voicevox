package voicevox

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/hajimehoshi/oto"
)

const (
	METAN = iota
	ZUNDA
	TSUMUGI
	HAU
	RITSU
	GENNO
	SHIRAGAMI
	AOYAMA
	HIMARI
	SORA
	MOCHIKO
	KENZAKI
	WHITECUL
	GOKI
	NO7
	CHIBIJII
	MICO
	SAYA
	NURSEROBOT
)

type Client struct {
	h          *http.Client
	Endpoint   string
	Speaker    int
	Style      int
	Speed      float64
	Intonation float64
	Volume     float64
	Pitch      float64
	Output     string
}

func New(addr string) *Client {
	c := http.DefaultClient

	temp, err := filepath.Abs(filepath.Clean("./output.wav"))
	if err != nil {
		return nil
	}
	return &Client{
		h:          c,
		Endpoint:   addr,
		Speaker:    0,
		Style:      0,
		Speed:      1.0,
		Intonation: 1.0,
		Volume:     1.0,
		Pitch:      0.0,
		Output:     temp,
	}
}

type Mora struct {
	Text            string   `json:"text"`
	Consonant       *string  `json:"consonant"`
	ConsonantLength *float64 `json:"consonant_length"`
	Vowel           string   `json:"vowel"`
	VowelLength     float64  `json:"vowel_length"`
	Pitch           float64  `json:"pitch"`
}

type AccentPhrases struct {
	Moras           []Mora `json:"moras"`
	Accent          int    `json:"accent"`
	PauseMora       *Mora  `json:"pause_mora"`
	IsInterrogative bool   `json:"is_interrogative"`
}

// {四国めたん 7ffcb7ce-00ec-4bdc-82cd-45a8889e43ff [{2 ノーマル} {0 あまあま} {6 ツンツン} {4 セクシー} {36 ささやき} {37 ヒソヒソ}] 0.14.1}
// {ずんだもん 388f246b-8c41-4ac1-8e2d-5d79f3ff56d9 [{3 ノーマル} {1 あまあま} {7 ツンツン} {5 セクシー} {22 ささやき} {38 ヒソヒソ}] 0.14.1}
// {春日部つむぎ 35b2c544-660e-401e-b503-0e14c635303a [{8 ノーマル}] 0.14.1}
// {雨晴はう 3474ee95-c274-47f9-aa1a-8322163d96f1 [{10 ノーマル}] 0.14.1}
// {波音リツ b1a81618-b27b-40d2-b0ea-27a9ad408c4b [{9 ノーマル}] 0.14.1}
// {玄野武宏 c30dc15a-0992-4f8d-8bb8-ad3b314e6a6f [{11 ノーマル} {39 喜び} {40 ツンギレ} {41 悲しみ}] 0.14.1}
// {白上虎太郎 e5020595-5c5d-4e87-b849-270a518d0dcf [{12 ふつう} {32 わーい} {33 びくびく} {34 おこ} {35 びえーん}] 0.14.1}
// {青山龍星 4f51116a-d9ee-4516-925d-21f183e2afad [{13 ノーマル}] 0.14.1}
// {冥鳴ひまり 8eaad775-3119-417e-8cf4-2a10bfd592c8 [{14 ノーマル}] 0.14.1}
// {九州そら 481fb609-6446-4870-9f46-90c4dd623403 [{16 ノーマル} {15 あまあま} {18 ツンツン} {17 セクシー} {19 ささやき}] 0.14.1}
// {もち子さん 9f3ee141-26ad-437e-97bd-d22298d02ad2 [{20 ノーマル}] 0.14.1}
// {剣崎雌雄 1a17ca16-7ee5-4ea5-b191-2f02ace24d21 [{21 ノーマル}] 0.14.1}
// {WhiteCUL 67d5d8da-acd7-4207-bb10-b5542d3a663b [{23 ノーマル} {24 たのしい} {25 かなしい} {26 びえーん}] 0.14.1}
// {後鬼 0f56c2f2-644c-49c9-8989-94e11f7129d0 [{27 人間ver.} {28 ぬいぐるみver.}] 0.14.1}
// {No.7 044830d2-f23b-44d6-ac0d-b5d733caa900 [{29 ノーマル} {30 アナウンス} {31 読み聞かせ}] 0.14.1}
// {ちび式じい 468b8e94-9da4-4f7a-8715-a22a48844f9e [{42 ノーマル}] 0.14.1}
// {櫻歌ミコ 0693554c-338e-4790-8982-b9c6d476dc69 [{43 ノーマル} {44 第二形態} {45 ロリ}] 0.14.1}
// {小夜/SAYO a8cc6d22-aad0-4ab8-bf1e-2f843924164a [{46 ノーマル}] 0.14.1}
// {ナースロボ＿タイプＴ 882a636f-3bac-431a-966d-c5e6bba9f949 [{47 ノーマル} {48 楽々} {49 恐怖} {50 内緒話}] 0.14.1}
type Speaker struct {
	Name        string   `json:"name"`
	SpeakerUUID string   `json:"speaker_uuid"`
	Styles      []Styles `json:"styles"`
	Version     string   `json:"version"`
}

type Styles struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (c *Client) GetSpeakers() ([]Speaker, error) {
	u, err := url.Parse(c.Endpoint + "/speakers")
	if err != nil {
		return nil, err
	}

	res, err := c.h.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	speakers := make([]Speaker, 0)
	if err := json.NewDecoder(res.Body).Decode(&speakers); err != nil {
		return nil, err
	}

	return speakers, nil
}

func (c *Client) GetQuery(id int, text string) (*Params, error) {
	u, err := url.Parse(c.Endpoint + "/audio_query")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("speaker", strconv.Itoa(id))
	q.Add("text", text)
	req.URL.RawQuery = q.Encode()

	res, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	params := new(Params)
	if err := json.NewDecoder(res.Body).Decode(&params); err != nil {
		return nil, err
	}

	return params, nil
}

type Params struct {
	AccentPhrases      []AccentPhrases `json:"accent_phrases"`
	SpeedScale         float64         `json:"speedScale"`
	PitchScale         float64         `json:"pitchScale"`
	IntonationScale    float64         `json:"intonationScale"`
	VolumeScale        float64         `json:"volumeScale"`
	PrePhonemeLength   float64         `json:"prePhonemeLength"`
	PostPhonemeLength  float64         `json:"postPhonemeLength"`
	OutputSamplingRate int             `json:"outputSamplingRate"`
	OutputStereo       bool            `json:"outputStereo"`
	Kana               string          `json:"kana"`
}

func (c *Client) Set(params *Params) {
	params.SpeedScale = c.Speed
	params.PitchScale = c.Pitch
	params.IntonationScale = c.Intonation
	params.VolumeScale = c.Volume
}

func (c *Client) Synth(id int, params *Params) ([]byte, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(c.Endpoint + "/synthesis")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "audio/wav")
	req.Header.Add("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("speaker", strconv.Itoa(id))

	req.URL.RawQuery = q.Encode()

	res, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buff := bytes.NewBuffer(nil)
	if _, err := io.Copy(buff, res.Body); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (c *Client) Speaking(params *Params, b []byte) error {
	// default output channel
	ch := 1
	if params.OutputStereo {
		ch = 2
	}

	ctx, err := oto.NewContext(params.OutputSamplingRate, ch, 2, 3200)
	if err != nil {
		return err
	}
	defer ctx.Close()

	p := ctx.NewPlayer()
	if _, err := io.Copy(p, bytes.NewReader(b)); err != nil {
		return err
	}
	if err := p.Close(); err != nil {
		return err
	}

	return nil
}
