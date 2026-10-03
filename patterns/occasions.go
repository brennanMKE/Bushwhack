package patterns

import "time"

// Occasion groups patterns by holiday or event. Window is the stretch of the
// year leading up to it, when it is "upcoming".
type Occasion struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Inclusive window as month/day; zero for Celebrations (always last).
	StartMonth, StartDay int `json:"-"`
	EndMonth, EndDay     int `json:"-"`
}

// Occasions in calendar order, Celebrations last. This one table decides
// what is "upcoming".
var Occasions = []Occasion{
	{"valentines", "Valentine's", "Hearts and arrows for gifts, boxes and cards.", 1, 1, 2, 14},
	{"spring", "Spring", "Eggs, flowers and bunnies for Easter and the first warm days.", 2, 15, 4, 30},
	{"july4", "July 4", "Stars and bursts for the Fourth.", 5, 1, 7, 4},
	{"halloween", "Halloween", "Jack-o'-lanterns, bats and ghosts. Back them with orange acrylic and light them up.", 9, 1, 10, 31},
	{"winter", "Winter holidays", "Snowflakes, trees and ornaments for the end of the year.", 11, 1, 12, 31},
	{"celebrations", "Celebrations", "Birthdays, weddings and anniversaries, any time of year.", 0, 0, 0, 0},
}

func (o Occasion) contains(t time.Time) bool {
	if o.StartMonth == 0 {
		return false
	}
	md := int(t.Month())*100 + t.Day()
	return md >= o.StartMonth*100+o.StartDay && md <= o.EndMonth*100+o.EndDay
}

// Upcoming returns the occasion whose window contains t. Between windows
// (Jul 5 to Aug 31) the next window to open wins.
func Upcoming(t time.Time) Occasion {
	for _, o := range Occasions {
		if o.contains(t) {
			return o
		}
	}
	md := int(t.Month())*100 + t.Day()
	for _, o := range Occasions {
		if o.StartMonth != 0 && o.StartMonth*100+o.StartDay > md {
			return o
		}
	}
	return Occasions[0]
}

// Ordered returns occasions starting with the upcoming one, then the rest in
// calendar order (wrapping), with Celebrations always last.
func Ordered(t time.Time) []Occasion {
	up := Upcoming(t)
	var dated []Occasion
	var last Occasion
	for _, o := range Occasions {
		if o.StartMonth == 0 {
			last = o
			continue
		}
		dated = append(dated, o)
	}
	i := 0
	for j, o := range dated {
		if o.Slug == up.Slug {
			i = j
		}
	}
	out := append(append([]Occasion{}, dated[i:]...), dated[:i]...)
	return append(out, last)
}

// Find returns the occasion with slug.
func Find(slug string) (Occasion, bool) {
	for _, o := range Occasions {
		if o.Slug == slug {
			return o, true
		}
	}
	return Occasion{}, false
}
