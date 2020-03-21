package common

import "github.com/matoous/mailback/internal/when/rules"

var All = []rules.Rule{
	SlashDMY(rules.Override),
}
