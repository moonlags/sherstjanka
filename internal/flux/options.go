package flux

type GenerationOptions struct {
	Seed              int
	Width             int
	Height            int
	NumInferenceSteps int
}

func (opts *GenerationOptions) populate() {
	if opts.NumInferenceSteps == 0 {
		opts.NumInferenceSteps = 4
	}

	if opts.Width == 0 {
		opts.Width = 1024
	}

	if opts.Height == 0 {
		opts.Height = 1024
	}
}
