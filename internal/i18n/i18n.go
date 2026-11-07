// Package i18n provides message translation backed by the .po catalogues
// embedded in the binary.
//
// Catalogues are read from po/*.po at build time, so the binary carries its
// translations and no longer needs cgo, libintl, or a locale directory
// installed next to it.
package i18n

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/lxc/incus/v7/po"
)

// entries caches the msgid -> msgstr table for the active locale.
var (
	once         sync.Once
	detectedLang string
	entries      map[string]string
)

// G returns the translation of msgid for the active locale.
//
// The lookup matches gettext: the locale named by LANGUAGE, LC_ALL,
// LC_MESSAGES or LANG, then its short language code, then the msgid itself.
// A msgid missing from the catalogue, or present with an empty msgstr, yields
// the msgid, so a catalogue lagging behind the source degrades to an
// untranslated string rather than crashing the process.
func G(msgid string) string {
	once.Do(load)

	translated, ok := entries[msgid]
	if ok && translated != "" {
		return translated
	}

	return msgid
}

// load resolves the active locale and parses its catalogue.
func load() {
	detectedLang = detectLanguage()
	if detectedLang == "" {
		return
	}

	for _, candidate := range catalogueNames(detectedLang) {
		data, err := po.Files.ReadFile(candidate + ".po")
		if err != nil {
			continue
		}

		entries = parsePo(string(data))

		return
	}
}

// catalogueNames lists the catalogue file names to try for a locale, most
// specific first: the locale itself, its alias, then the same two for the
// bare language code.
//
// Trying the alias before the short code matters for Chinese. A zh_CN locale
// must reach zh_Hans, and the short code "zh" has no catalogue, so putting
// the short code first would skip straight to English.
func catalogueNames(lang string) []string {
	var names []string

	add := func(name string) {
		if name == "" {
			return
		}

		for _, existing := range names {
			if existing == name {
				return
			}
		}

		names = append(names, name)
	}

	add(lang)
	add(localeAlias[lang])

	short := shortCode(lang)
	add(short)
	add(localeAlias[short])

	return names
}

// localeAlias maps a locale to the catalogue that serves it when the two are
// named differently.
//
// Chinese catalogues are named by script, zh_Hans and zh_Hant, while locales
// are named by region. Without this every zh_* locale falls back to English,
// which is what incus did before the catalogues were embedded: the region
// names never matched a file. The region-to-script assignment is CLDR's.
//
// Norwegian: nb_NO is Bokmål, and the legacy no_NO alias means the same
// thing. Nynorsk (nn_*) is a separate written standard with no catalogue
// here, so it is deliberately absent and falls back to English rather than
// being served Bokmål.
var localeAlias = map[string]string{
	"zh_CN": "zh_Hans",
	"zh_SG": "zh_Hans",
	"zh_MY": "zh_Hans",
	"zh_TW": "zh_Hant",
	"zh_HK": "zh_Hant",
	"zh_MO": "zh_Hant",
	"no":    "nb",
	"no_NO": "nb_NO",
}

// shortCode reduces "zh_Hans" to "zh", or returns "" when there is no "_".
func shortCode(lang string) string {
	i := strings.IndexByte(lang, '_')
	if i > 0 {
		return lang[:i]
	}

	return ""
}

// detectLanguage returns the locale name to look up, or "" for none.
//
// LANGUAGE may hold a colon-separated list, and every variable may carry an
// encoding (.UTF-8) and a modifier (@euro), both stripped before the
// catalogue file name is formed. C and POSIX mean "no translation".
func detectLanguage() string {
	for _, name := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		val := os.Getenv(name)
		if val == "" || val == "C" || val == "POSIX" {
			continue
		}

		i := strings.IndexByte(val, ':')
		if i >= 0 {
			val = val[:i]
		}

		i = strings.IndexByte(val, '.')
		if i >= 0 {
			val = val[:i]
		}

		i = strings.IndexByte(val, '@')
		if i >= 0 {
			val = val[:i]
		}

		if val != "" {
			return val
		}
	}

	return ""
}

// parsePo reads a catalogue into a msgid -> msgstr map.
//
// Entries with an empty msgstr are skipped: they are untranslated and G()
// already returns the msgid for them. Plural forms are not represented,
// because incus does not call ngettext.
func parsePo(data string) map[string]string {
	out := make(map[string]string)

	var msgid, msgstr string

	var inID, inStr bool

	flush := func() {
		if msgid != "" && msgstr != "" {
			out[msgid] = msgstr
		}

		msgid, msgstr = "", ""
		inID, inStr = false, false
	}

	scanner := bufio.NewScanner(strings.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "#"):
			// comment, reference or flag line
		case strings.HasPrefix(line, `msgid `):
			if msgid != "" {
				flush()
			}

			msgid = unquotePo(strings.TrimPrefix(line, `msgid `))
			inID, inStr = true, false
		case strings.HasPrefix(line, `msgstr `):
			msgstr = unquotePo(strings.TrimPrefix(line, `msgstr `))
			inID, inStr = false, true
		case inID:
			msgid += unquotePo(line)
		case inStr:
			msgstr += unquotePo(line)
		}
	}

	flush()

	return out
}

// unquotePo strips one pair of surrounding quotes and unescapes the result.
//
// Trimming exactly one leading and one trailing quote matters: a value ending
// in an escaped quote, like "say \"hi\"", would be mangled by strings.Trim.
func unquotePo(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	return unescapePo(s)
}

// unescapePo decodes the C-style escapes used in .po strings.
//
// Backslash must be decoded last, otherwise "\\n" would turn into a newline
// instead of a literal backslash followed by the letter n.
func unescapePo(s string) string {
	s = strings.ReplaceAll(s, `\\`, "\x00")
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\n`, "\n")

	return strings.ReplaceAll(s, "\x00", `\`)
}
