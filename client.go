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

func New() *Client {
	c := http.DefaultClient

	temp, err := filepath.Abs(filepath.Clean("./output.wav"))
	if err != nil {
		return nil
	}
	return &Client{
		h:          c,
		Endpoint:   "http://localhost:50021",
		Speaker:    0,
		Style:      0,
		Speed:      1.0,
		Intonation: 1.0,
		Volume:     1.0,
		Pitch:      0.0,
		Output:     temp,
	}
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
	b, err := json.MarshalIndent(params, "", "  ")
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
