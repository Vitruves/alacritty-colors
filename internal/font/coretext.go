package font

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// On macOS, Alacritty resolves fonts through Core Text, not fontconfig.
//
// That distinction is the whole reason this file exists. fc-list is a Homebrew
// extra that many Macs do not have at all, and where it is installed it reports
// families and style names that Core Text does not always agree with — it lists
// private system faces such as ".SF NS Mono", and it can give style names that
// Alacritty then fails to match. Offering those is what produced "unable to
// load font": the font was listed, so it looked installed, but the name handed
// to Alacritty was not one Core Text answers to.
//
// Asking the same font database Alacritty will ask removes the guesswork. The
// script below walks every family, keeps those with at least one monospaced
// face, and reports the style names exactly as Core Text spells them.

const coreTextScript = `
ObjC.import('AppKit');
var fm = $.NSFontManager.sharedFontManager;
var fams = fm.availableFontFamilies;
var out = [];
for (var i = 0; i < fams.count; i++) {
  var fam = ObjC.unwrap(fams.objectAtIndex(i));
  if (fam.charAt(0) === '.') continue;
  var members = fm.availableMembersOfFontFamily($(fam));
  if (!members) continue;
  var styles = [], mono = false;
  for (var j = 0; j < members.count; j++) {
    var m = members.objectAtIndex(j);
    var style = ObjC.unwrap(m.objectAtIndex(1));
    var traits = ObjC.unwrap(m.objectAtIndex(3));
    if ((traits & 1024) !== 0) mono = true;
    styles.push(style);
  }
  if (mono) out.push(fam + '\t' + styles.join('|'));
}
out.join('\n');
`

// NSFontMonoSpaceTrait is bit 10 of the trait mask, tested in the script above.

// coreTextFamilies returns monospaced families mapped to their style names, as
// Core Text spells both. The scan takes roughly a third of a second, which is
// why the font panel runs it behind a spinner.
func coreTextFamilies() (map[string][]string, error) {
	cmd := exec.Command("osascript", "-l", "JavaScript", "-e", coreTextScript)
	cmd.Env = append(os.Environ(), "LC_ALL=en_US.UTF-8")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not read the Core Text font list: %w", err)
	}

	families := make(map[string][]string)
	for _, line := range strings.Split(string(output), "\n") {
		family, styleList, found := strings.Cut(strings.TrimSpace(line), "\t")
		if !found || family == "" {
			continue
		}

		var styles []string
		for _, style := range strings.Split(styleList, "|") {
			if style = strings.TrimSpace(style); style != "" {
				styles = append(styles, style)
			}
		}
		if len(styles) == 0 {
			styles = []string{"Regular"}
		}
		families[family] = styles
	}

	if len(families) == 0 {
		return nil, fmt.Errorf("Core Text reported no monospaced families")
	}
	return families, nil
}

// sortedKeys returns the family names in display order.
func sortedKeys(families map[string][]string) []string {
	names := make([]string, 0, len(families))
	for name := range families {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
