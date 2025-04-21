package repeatable

import "time"

func DoWithTries(fn func() error, attemtps int, delayFragment time.Duration) (err error) {
	delay := delayFragment
	var delayIncreaseFactor int64 = 1
	for attemtps > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attemtps--
			delayIncreaseFactor++
			delay = time.Duration(delay.Nanoseconds() * delayIncreaseFactor)
			continue
		}

		return nil
	}

	return
}
