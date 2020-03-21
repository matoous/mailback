package en

import (
	"regexp"
	"strings"
	"time"

	"github.com/rickb777/date/period"

	"github.com/matoous/mailback/internal/when/rules"
)

func CasualPeriod(s rules.Strategy) rules.Rule {
	return &rules.F{
		RegExp: regexp.MustCompile(`(?i)(?:\W|^)(daily|weekly|monthly|quarterly|yearly|annually)(?:\W|$)`),
		Applier: func(m *rules.Match, c *rules.Context, o *rules.Options, ref time.Time) (bool, error) {
			lower := strings.ToLower(strings.TrimSpace(m.String()))

			switch {
			case strings.Contains(lower, "daily"):
				p := period.NewYMD(0, 0, 1)
				c.Period = &p
			case strings.Contains(lower, "weekly"):
				p := period.NewYMD(0, 0, 7)
				c.Period = &p
			case strings.Contains(lower, "monthly"):
				p := period.NewYMD(0, 1, 0)
				c.Period = &p
			case strings.Contains(lower, "quarterly"):
				p := period.NewYMD(0, 3, 0)
				c.Period = &p
			case strings.Contains(lower, "yearly"):
			case strings.Contains(lower, "annually"):
				p := period.NewYMD(1, 0, 0)
				c.Period = &p
			}

			return true, nil
		},
	}
}
