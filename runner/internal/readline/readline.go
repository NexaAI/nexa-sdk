package readline

import ()

type Config struct {
	Prompt string
}

type Readline struct {
	term *Terminal
	buf  *Buffer
}

func New(config *Config) (*Readline, error) {
	term, err := NewTerminal()
	if err != nil {
		return nil, err
	}

	buf := NewBuffer()
	buf.prompt = config.Prompt

	return &Readline{
		term: term,
		buf:  buf,
	}, nil
}

func (r *Readline) Read() (string, error) {
	return r.buf.Read()
}

func (r *Readline) Close() error {
	return r.term.Close()
}
