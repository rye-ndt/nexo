package harness_helper

type Unwind struct {
	steps []func()
}

func (u *Unwind) Push(step func()) {
	u.steps = append(u.steps, step)
}

func (u *Unwind) Done() {
	u.steps = nil
}

func (u *Unwind) Run() {
	for i := len(u.steps) - 1; i >= 0; i-- {
		u.steps[i]()
	}

	u.steps = nil
}
