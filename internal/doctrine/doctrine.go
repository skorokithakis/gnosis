package doctrine

import _ "embed"

//go:embed help.txt
var Help string

//go:embed plan.txt
var Plan string

//go:embed review.txt
var Review string

//go:embed commands.txt
var Commands string
