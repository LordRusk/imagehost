package main

import (
	"strings"
	"testing"
	"time"
)

var testFile = "/tmp/" + strings.ReplaceAll(strings.ReplaceAll(time.Now().Format(time.UnixDate), "  ", " "), " ", "-")

func TestLogFile(t *testing.T) {
	lf := newLogFile(testFile)
	if err := lf.get(); err == nil { // file should not exit
		t.Errorf("'%s' should not exist: %s\n", testFile, err)
	}

	lf.records = map[string]imgRecord{
		"1": {
			Name:  "test record #1",
			Added: time.Now(),
			Path:  "“Prophet, i.E a critic and saturist of the moment” - Nietzche",
		},
		"2": {
			Name:  "test record #2",
			Added: time.Now(),
			Path:  "“If nature has made any one thing less susceptible than all others of exclusive property, it is the action of the thinking power called an idea, which an individual may exclusively possess as long as he keeps it to himself; but the moment it is divulged, it forces itself into the possession of everyone, and the receiver cannot dispossess himself of it. Its peculiar character, too, is that no one possesses the less, because every other possesses the whole of it. He who receives an idea from me, receives instruction himself without lessening mine; as he who lights his taper at mine, receives light without darkening me. That ideas should freely spread from one to another over the globe, for the moral and mutual instruction of man, and improvement of his condition, seems to have been peculiarly and benevolently designed by nature, when she made them, like fire, expansible over all space, without lessening their density in any point, and like the air in which we breathe, move, and have our physical being, incapable of confinement or exclusive appropriation. Inventions then cannot, in nature, be a subject of property.” - Thomas Jefferson",
		},
		"3": {
			Name:  "test record #3",
			Added: time.Now(),
			Path:  "“Copying isn’t theft, and it isn’t piracy. It’s what we did for millennia until the invention of copyright, and we can do it again, if we don’t hobble ourselves with the antiquated remnants of a censorship system from the sixteenth century.” — Karl Fogel",
		},
	}

	if err := lf.save(); err != nil {
		t.Errorf("Failed to save: %s\n", err)
	}

	if err := lf.get(); err != nil {
		t.Errorf("Failed to get saved logs: %s\n", err)
	}

	lf = newLogFile(testFile)
	if err := lf.get(); err != nil {
		t.Errorf("Failed to get saved logs: %s\n", err)
	}
}
