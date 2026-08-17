package harness_helper

// TokenCounts is one reading of what a turn has cost: Billed is what the agent wrote,
// Input is prompt read at full price, Cached is prompt read back out of the vendor's
// cache.
type TokenCounts struct {
	Billed int
	Input  int
	Cached int
}

func (t TokenCounts) plus(other TokenCounts) TokenCounts {
	return TokenCounts{
		Billed: t.Billed + other.Billed,
		Input:  t.Input + other.Input,
		Cached: t.Cached + other.Cached,
	}
}

func (t TokenCounts) highest(other TokenCounts) TokenCounts {
	return TokenCounts{
		Billed: max(t.Billed, other.Billed),
		Input:  max(t.Input, other.Input),
		Cached: max(t.Cached, other.Cached),
	}
}

// TurnMeter turns a stream of usage readings into a running total. Every harness
// streams the same message repeatedly as it grows, so both of them need this and both
// of them need it to behave identically — a counter that rolled over on a different
// event in one harness would bill that vendor differently for the same work.
type TurnMeter struct {
	settled  TokenCounts
	msgID    string
	inFlight TokenCounts
}

// Observe takes one reading of the message named by id and answers what has been spent
// so far. The same message is reported again as it streams, so its counts only join the
// running total once the next message starts. A reading with no id cannot be matched
// that way, so it counts as a turn of its own — and leaves the named message in flight
// alone, which would otherwise be billed again when it streams on.
func (m *TurnMeter) Observe(id string, reported TokenCounts) TokenCounts {
	if id == "" {
		m.settled = m.settled.plus(reported)

		return m.settled.plus(m.inFlight)
	}

	if id != m.msgID {
		m.settled = m.settled.plus(m.inFlight)
		m.msgID = id
		m.inFlight = TokenCounts{}
	}

	m.inFlight = m.inFlight.highest(reported)

	return m.settled.plus(m.inFlight)
}
