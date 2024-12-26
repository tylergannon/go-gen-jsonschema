package testapp3validate

import (
	_ "github.com/dave/dst/decorator"
	_ "github.com/santhosh-tekuri/jsonschema"
	_ "github.com/tylergannon/go-gen-jsonschema"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/go-gen-jsonschema-testapp/llmfriendlytimepkg3"
	_ "github.com/tylergannon/structtag"
	_ "golang.org/x/tools/go/packages"
	"time"
)

//go:generate go run github.com/tylergannon/go-gen-jsonschema/gen-jsonschema/ -type MovieCharacter -pretty -validate

type (
	LLMFriendlyTime time.Time

	MovieWood string
	Alignment string

	ImportantDate struct {
		When         LLMFriendlyTime `json:"when"`
		WhatHappened string          `json:"whatHappened"`
		WithWhom     string          `json:"withWhom"`
	}

	Weapon string
	// MovieCharacter is a character not just in a story, but particularly in a
	// movie.  NOTE that this is not for actors, but for characters in movies.
	// MUST BE COMPLETELY FICTITIONAL.  DO NOT COPY REAL CHARACTERS FROM REAL
	// MOVIES!  YOU WILL BE SUED FOR COPYRIGHT INFRINGEMENT.
	MovieCharacter struct {
		// Name is the character's legal name.
		Name string `json:"name"`
		// Nicknames is what they're called by their friends, and also by their enemies or detractors
		Nicknames []string `json:"nicknames"`
		// Location is which movie-making industry they come from in.
		Location  MovieWood `json:"location"`
		Alignment Alignment `json:"alignment"`
		// FavoriteWeapons choose up to five
		FavoriteWeapons []Weapon `json:"favoriteWeapons"`

		// DateOfBirth is when they were born.
		DateOfBirth LLMFriendlyTime `json:"dateOfBirth"`

		// BackStory describes how this character came to be who they are
		BackStory string `json:"backStory"`

		// ImportantDates is a list of major events that happen in the person's life.
		// List at least three.
		ImportantDates []ImportantDate `json:"importantDates"`
	}
)

const (
	Good    Alignment = "good"
	Neutral Alignment = "neutral"
	Evil    Alignment = "evil"
)

const (
	HollyWood  MovieWood = "HollyWood"  // United States (English-language films)
	BollyWood  MovieWood = "BollyWood"  // India (Hindi-language films)
	TollyWood  MovieWood = "TollyWood"  // India (Telugu-language films and Bengali-language films, used in two contexts)
	KollyWood  MovieWood = "KollyWood"  // India (Tamil-language films)
	SandalWood MovieWood = "SandalWood" // India (Kannada-language films)
	MollyWood  MovieWood = "MollyWood"  // India (Malayalam-language films)
	LollyWood  MovieWood = "LollyWood"  // Pakistan (Urdu- and Punjabi-language films)
	Chollywood MovieWood = "Chollywood" // Peru
	Nollywood  MovieWood = "Nollywood"  // Nigeria (English- and indigenous-language films)
	HallyuWood MovieWood = "HallyuWood" // South Korea (Korean Wave cinema)
	Pollywood  MovieWood = "Pollywood"  // Punjab region (Punjabi-language films, used in India and Pakistan)
	DhallyWood MovieWood = "DhallyWood" // Bangladesh (Bengali-language films)
)

const (
	Sword                 Weapon = "Sword"                 // A classic melee weapon.
	Bow                   Weapon = "Bow"                   // Traditional ranged weapon with arrows.
	Spear                 Weapon = "Spear"                 // Long polearm weapon.
	SaturdayNightSpecial  Weapon = "SaturdayNightSpecial"  // Small, inexpensive handgun.
	Grenade               Weapon = "Grenade"               // Explosive throwable weapon.
	Bazooka               Weapon = "Bazooka"               // Shoulder-fired rocket launcher.
	SodaStrawSpitballs    Weapon = "SodaStrawSpitballs"    // Silly, improvised ranged weapon.
	OceanEvaporatingLaser Weapon = "OceanEvaporatingLaser" // Extreme, sci-fi superweapon.
	Crossbow              Weapon = "Crossbow"              // Bow-like mechanical ranged weapon.
	Flamethrower          Weapon = "Flamethrower"          // Weapon emitting flames.
	Katana                Weapon = "Katana"                // Japanese curved sword.
	Chainsaw              Weapon = "Chainsaw"              // Melee weapon of destruction.
	Nunchaku              Weapon = "Nunchaku"              // Traditional martial arts weapon.
	Slingshot             Weapon = "Slingshot"             // Simple ranged weapon.
	Blowgun               Weapon = "Blowgun"               // Simple tube-based ranged weapon.
	PaperAirplaneCutouts  Weapon = "PaperAirplaneCutouts"  // Silly improvised "weapon."
	TacticalNuke          Weapon = "TacticalNuke"          // Compact nuclear weapon.
	LaserPointerBlinder   Weapon = "LaserPointerBlinder"   // Blinding weapon using lasers.
	BallisticMissile      Weapon = "BallisticMissile"      // Strategic weapon for massive destruction.
	Boomerang             Weapon = "Boomerang"             // Ranged weapon that can return to the thrower.
	RocketLauncher        Weapon = "RocketLauncher"        // Weapon for launching rockets.
	FryingPan             Weapon = "FryingPan"             // Silly yet effective improvised melee weapon.
	PocketKnife           Weapon = "PocketKnife"           // Small, versatile blade.
	HarpoonGun            Weapon = "HarpoonGun"            // Underwater projectile weapon.
	GravityHammer         Weapon = "GravityHammer"         // Extreme, sci-fi melee weapon.
	Peashooter            Weapon = "Peashooter"            // Comically weak ranged weapon.
	GatlingGun            Weapon = "GatlingGun"            // Rapid-fire gun.
	PlasmaCannon          Weapon = "PlasmaCannon"          // Sci-fi energy-based weapon.
	SnowballLauncher      Weapon = "SnowballLauncher"      // Silly winter-themed ranged weapon.
	MorningStar           Weapon = "MorningStar"           // Spiked mace.
	LaserSword            Weapon = "LaserSword"            // Sci-fi melee weapon akin to a lightsaber.
)

func TimeAgoToLLMFriendlyTime(t llmfriendlytimepkg3.TimeAgo) (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now().Add(-llmfriendlytimepkg3.ToDuration(t.Unit, t.Quantity))), nil
}

func FromNowToLLMFriendlyTime(t llmfriendlytimepkg3.TimeFromNow) (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now().Add(llmfriendlytimepkg3.ToDuration(t.Unit, t.Value))), nil
}

func ActualTimeToLLMFriendlyTime(t llmfriendlytimepkg3.ActualTime) (LLMFriendlyTime, error) {
	_t, err := t.ToTime()
	return LLMFriendlyTime(_t), err
}

func NowToLLMFriendlyTime(t llmfriendlytimepkg3.Now) (LLMFriendlyTime, error) {
	_t, err := t.ToTime()
	return LLMFriendlyTime(_t), err
}
func BeginningOfTimeToLLMFriendlyTime(t llmfriendlytimepkg3.BeginningOfTime) (LLMFriendlyTime, error) {
	_t, err := t.ToTime()
	return LLMFriendlyTime(_t), err
}

var _ = jsonschema.SetTypeAlternative[LLMFriendlyTime](
	// For referencing a time in the past using relative units
	jsonschema.Alt("timeAgo", TimeAgoToLLMFriendlyTime),
	// For referencing a time in the future using relative units
	jsonschema.Alt("timeFromNow", FromNowToLLMFriendlyTime),
	// When given an actual time. Must be valid RFC3339 time format
	jsonschema.Alt("actualTime", ActualTimeToLLMFriendlyTime),
	// To refer to the present moment
	jsonschema.Alt("now", NowToLLMFriendlyTime),
	// To reference all of history
	jsonschema.Alt("beginningOfTime", BeginningOfTimeToLLMFriendlyTime),
	jsonschema.Alt("nearestDay", NearestDayToTime),
	jsonschema.Alt("nearestDate", NearestDateToTime),
)
