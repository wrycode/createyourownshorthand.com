package main

import (
	"net/http"
	"io"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"image"
	"github.com/beevik/prefixtree"
	"fmt"
	"encoding/json"
	"image/color"
	"github.com/beevik/etree"
	"log"
	"strconv"
	"strings"
	"net/url"
	"html"
	"errors"
	"embed"
	"math"
	"unicode"
)

// Options for rendering
type Options struct {
	Debug bool `json:"Debug"`
	Draw_bounding_boxes bool `json:"Draw_bounding_boxes"`
	Form_stroke_width float64 `json:"Form_stroke_width"`
	Custom_script_svg_value  string  `json:"Custom_script_svg_value"`
	Builtin_script_name string  `json:"Builtin_script_name"`
	Language_code string `json:"Language_code"`
	Image_width float64 `json:"Image_width"`
	Input_text string `json:"Input_text"`
	Space_between_metaforms float64 `json:"Space_between_metaforms"`
	Margin float64 `json:"Margin"`
	Space_between_lines float64 `json:"Space_between_lines"`
}

func LoadIPADict(lang string, log *log.Logger) (map[string]string, error) {
	type IPAJson map[string][]map[string]string
	var jsonDict IPAJson

	resp, err := http.Get(fmt.Sprintf("https://f005.backblazeb2.com/file/ipa-dict/%s.json", lang))
	if err != nil {
		log.Println("Error downloading IPA dictionary: ", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading IPA dictionary: " ,err)
	}
	
	err = json.Unmarshal(body, &jsonDict)
	if err != nil {
		log.Println("Error unmarshaling IPA dictionary: " ,err)
		return nil, err
	}
	return jsonDict[lang][0], err
}

// Convert SVG path description from Inkscape into path description for canvas library
// TODO: May need to improve this function in the future! Have not tested extensively.
func SVG_path_to_canvas(svgpath string) string {
	split := strings.Split(svgpath, " ")
	// remove the first 'move' command to make these paths relative
	split = split[2:]
	switch split[0] {
	case "m", "l", "h", "v", "c", "s", "q", "t", "a", "z":		// command is already present, don't need to do anything
	default:
		split = append([]string{"l"}, split...) // need to add the lineto command
	}
	for i, str := range split {
		split[i] = reverseYAxis(str)
	}
	return strings.Join(split, " ")
}

func reverseYAxis(coord string) string {
	points := strings.Split(coord, ",")
	if len(points) != 2 {
		return coord
	}
	y, err := strconv.ParseFloat(points[1], 64)
	if err != nil {
		return coord
	}
	y = -y
	return fmt.Sprintf("%s,%f", points[0], y)
}

/* A handwriting system definition

   SubForms are Tokens representing a sequence of one or more IPA
   characters. They are stored in a prefix tree so the parser can
   easily select the largest subform possible while scanning the IPA
   text.

   Logos are Tokens representing a full word (Logogram) or more than
   one word (phrase). They are also stored in a prefix tree.

*/

type Script struct {
	SubForms *prefixtree.Tree
	Logos *prefixtree.Tree
}

// Load dictionary of subforms and logograms
func load_script(script string) *Script {

	// TODO: need to validate the script to alert the user if an
	// IPA sequence (or word or phrase) is mapped to multiple
	// tokens. The other way around is fine, however: one token
	// can represent multiple IPA sequences (e.g. 'w' and 'hw' are
	// simplified to one symbol) or words or phrases.

	subforms := prefixtree.New()
	logograms := prefixtree.New()

	doc := etree.NewDocument()
	if err := doc.ReadFromString(script); err != nil {
		log.Fatalf("Failed to parse document: %v", err)
	}

	root := doc.SelectElement("svg")
	for _, path_element := range root.FindElements("//path") {
		label := path_element.SelectAttrValue("inkscape:label", "unknown")

		values, err := url.ParseQuery(html.UnescapeString(label))
		if err != nil {
			panic(err)
		}

		if ipa_field_val := values.Get("IPA"); ipa_field_val != "" {
			canvas_path := SVG_path_to_canvas(path_element.SelectAttr("d").Value)
			for _, ipa_sequence := range strings.Split(ipa_field_val, ",") {
				token := Token {
					Name: ipa_sequence,
						Path: canvas_path,
					}
				subforms.Add(ipa_sequence, token)
			}
		}

		if logo_field_val := values.Get("logo"); logo_field_val != "" {
			canvas_path := SVG_path_to_canvas(path_element.SelectAttr("d").Value)
			for _, logo_str := range strings.Split(logo_field_val, ",") {
				token := Token {
					Name: logo_str,
						Path: canvas_path,
					}
				logograms.Add(logo_str, token)
			}
		}
	}

	return &Script{
		SubForms: subforms,
		Logos: logograms,
	}
}

// Return a string, placing a space before and after certain
// characters (including NEWLINES!). The reason is that newlines and
// punctuation should be treated as their own metaform.
func normalizePunctuation(input string) string {
	punctuations := ".,;:!?\n"
	for _, punc := range punctuations {
		// add spaces around punctuation only if they don't exist yet
		input = strings.ReplaceAll(input, string(punc), " "+string(punc)+" ")
	}
	// Remove extra spaces
	input = strings.Join(customFields(input), " ")
	return input
}

/* a Token is an indivisible 'unit' of a handwritten document in a
 specific script. The input text is parsed into a sequence of
 Tokens. A Token can be one of the following 3 things:

   - alphabetical subform, representing a sequence of one or more IPA characters
   - logogram, representing a full word or phrase in the target language
   - whitespace, punctuation mark or other unknown symbol or character

   Naming: A subform is named the sequence of IPA characters it
   represents.  A logogram is named the word or phrase it represents
   in the target language. All other Tokens represent a single
   character (whitespace or otherwise) and are named that character.

   If the Token is defined in the current Script, Path contains a
   Canvas path, otherwise it is empty.
*/

type Token struct {
	Name  string
	Path  string
}

func (t Token) String() string {
	return fmt.Sprintf("%v Path: %v", t.Name, t.Path)
}

/* A Metaform is a sequence of one or more Tokens that are drawn
   connected together in a 'word'. A Metaform can be a single Token,
   like "," (comma), logograms, like "my" or "how are you", or a
   sequence of Tokens, for instance [r·aɪ·t·ɪŋ], representing the word
   "writing". ` ` (space) is the only token that is not converted to a
   Metaform, because all metaforms are assumed to be separated by a
   space. Spacing drawn between metaforms is therefore handled by the
   rendering code. Newlines are separated by spaces and converted to
   metaforms like other punctuation, then the rendering code handles
   creating a new line. All other whitespace chars (like tabs) are
   ignored.

   Metaforms also contain extra information for debugging and
   rendering.
*/
type Metaform struct {
	Name string // The string of characters represented by the Metaform
	Tokens []Token		// must have at least 1
	contains_out_of_script_characters bool
	Img image.Image
}

func (m Metaform) String() string {
	var tokens []string
	for _, token := range m.Tokens {
		tokens = append(tokens, token.Name)
	}
	return strings.Join(tokens, "·")
}

// Render the form in a first pass and store it in m.Img, using the
// smallest image size that fits the form
func (m *Metaform) renderForm(o Options, log *log.Logger) *Metaform {
	c := canvas.New(5,5) // Arbitrary width and height - will be cut to size later
	ctx := canvas.NewContext(c)
	var Transparent = color.RGBA{0x00, 0x00, 0x00, 0x00}
	ctx.SetFillColor(Transparent)
	ctx.SetStrokeColor(canvas.Black)
	ctx.SetStrokeWidth(o.Form_stroke_width)

	// Create a single path to work with
	pos := canvas.Point{X: 0, Y: 0}
	path, _ := canvas.ParseSVGPath("")

	for _, t := range m.Tokens {
		if t.Path != "" {
			newPath, err := canvas.ParseSVGPath(t.Path)
			if err != nil {
				log.Println("Error parsing path: for ",m.Name, t.Name, err)
			}
			newPath = newPath.Translate(pos.X, pos.Y)
			path = path.Join(newPath)
			pos.X = path.Pos().X
			pos.Y = path.Pos().Y
		}
	}

	ctx.DrawPath(0, 0, path)
	bounding_box := path.Bounds()

	if o.Draw_bounding_boxes {
		ctx.SetStrokeColor(canvas.Red)
		ctx.SetStrokeWidth(0.1)
		ctx.DrawPath(0, 0, bounding_box.ToPath())
	}

	// Crop the image to the path
	c.Fit(1)
	ctx.SetCoordRect(bounding_box, bounding_box.W, bounding_box.H)
	m.Img = rasterizer.Draw(c, canvas.DefaultResolution, nil)
	return m
}

//go:embed font/*
var fonts embed.FS

// Redraw the form in m.Img into the center of a new image of the
// given height, adding debugging information if specified
func (m *Metaform) centerForm(new_height float64, o Options, face *canvas.FontFace) *Metaform {
	new_height =  new_height / canvas.DefaultResolution.DPMM()
	form_height := float64(m.Img.Bounds().Dy()) / (canvas.DefaultResolution.DPMM())
	form_width := float64(m.Img.Bounds().Dx()) / (canvas.DefaultResolution.DPMM()) // won't be changed

	if o.Debug {
		new_height += 10

	}
	c := canvas.New(form_width, new_height)

	ctx := canvas.NewContext(c)
	if o.Debug {
		ctx.DrawImage(0,(new_height - form_height + 10) / 2 , m.Img, canvas.DefaultResolution)
	} else {
		ctx.DrawImage(0,(new_height - form_height) / 2 , m.Img, canvas.DefaultResolution)
	}

	var Transparent = color.RGBA{0x00, 0x00, 0x00, 0x00}


	if o.Draw_bounding_boxes {
		ctx.SetFillColor(Transparent)
		ctx.SetStrokeColor(canvas.Blue)
		ctx.SetStrokeWidth(0.05)
		border := canvas.Rect{0, 0,form_width, new_height }
		ctx.DrawPath(0, 0, border.ToPath())
	}

	if o.Debug {

		debug_text := m.Name + " (" + m.String() + ")"
		if m.Name == "\n" {
			debug_text = "¶"
		}
		ctx.DrawText(0, 10, canvas.NewTextBox(face, debug_text, form_width, 10, canvas.Left, canvas.Top, 0.0, 0.0))
	}

	m.Img = rasterizer.Draw(c, canvas.DefaultResolution, nil)
	return m
}

// A document is a sequence of Metaforms to be rendered.
type Document struct {
	Metaforms []*Metaform
}

// like strings.Fields but don't treat newlines as whitespace
func customFields(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) && r != '\n'
	})
}

// returns a Document to be rendered
func Parse(input string, lcode string, script *Script, log *log.Logger) Document {
	// detach punctuation and newlines from words
	input = normalizePunctuation(input)

	input = strings.ToLower(input)

	ipa, err := LoadIPADict(lcode, log)
	if err != nil {
		log.Fatal(err)
	}

	words := customFields(input)

	doc := Document{
		Metaforms: make([]*Metaform, 0, len(words)*2),
	}

	// Loop through words, converting to logograms or IPA and appending to the document
	for i := 0; i < len(words); {
		word := words[i]

		// First check if the word is found in the logogram prefix tree
		val, err := script.Logos.FindValue(word)
		var matched_phrase_or_logo string // empty unless we find a matching sequence of one or more words in the script.Logos prefix tree

		if err == nil {
			logo := val.(Token)
			next_word_pos := i + 1 // Only used to update i later if we find a matching phrase or logo

			if logo.Name == word  { // exact match
				matched_phrase_or_logo = word
			}

			// Keep checking words until the sequence of
			// words isn't found in the prefix tree
			for j := i + 2; j < len(words); j++ {
				next_phrase := strings.Join(words[i:j], " ")
				next_val, err := script.Logos.FindValue(next_phrase)

				if err == nil {
					next_logo := next_val.(Token)
					if next_logo.Name == next_phrase { // exact match
						logo = next_logo
						matched_phrase_or_logo = next_phrase
						next_word_pos = j
					}
				} else if errors.Is(err, prefixtree.ErrPrefixAmbiguous) {
					// continue searching for a phrase but don't save the current one
					// because we don't have a new match
				} else { // No match for the current word at j; time to save the logogram
					break
				}
			}

			if matched_phrase_or_logo != "" {
				i = next_word_pos
				metaform := Metaform{
					Tokens:        []Token{logo},
					Name: matched_phrase_or_logo,
				}
				doc.Metaforms = append(doc.Metaforms, &metaform)
				continue
			}
		}

		// If we got this far in the loop, then we can assume
		// no phrase or logogram was found. Convert the word
		// to IPA:
		if replacement, exists := ipa[word]; exists {
			// when there are several possible
			// pronunciations, right now we just select
			// the first option
			first_option := strings.SplitN(replacement, ",", 2)[0]

			// strip forward slashes and accent characters
			// we're not using right now
			first_option = strings.ReplaceAll(first_option, "/", "")
			first_option = strings.ReplaceAll(first_option, "ˈ", "")
			first_option = strings.ReplaceAll(first_option, "ˌ", "")
			words[i] = first_option
		}

		// Now convert the IPA characters into subform tokens

		// easier to index and loop through, we just
		// have to cast back into string when we
		// search the subforms prefix tree
		chars := []rune(words[i])

		end := len(chars)
		current_char := 0

		metaform := Metaform{
			Tokens:        []Token{},
			Name: word,
		}

		for current_char < end {
			seq_end := current_char + 1
			form_key := string(chars[current_char:seq_end]) // default to one character form
			val, err := script.SubForms.FindValue(form_key)

			if errors.Is(err, prefixtree.ErrPrefixNotFound) {
				// character is not defined in the script, so we'll just store it as a pathless Token and move on
				token := Token {
					Name: form_key,
					}
				metaform.Tokens = append(metaform.Tokens, token)
				metaform.contains_out_of_script_characters = true
				log.Printf("Note: Metaform %v contains out-of-script character %v", metaform.Name, token.Name)
			}

			for seq_end < end {
				seq_end+= 1
				new_form_key := string(chars[current_char:seq_end])
				new_val, err := script.SubForms.FindValue(new_form_key)

				if err == nil { // new, longer form found
					if new_val.(Token).Name == new_form_key  { // exact match
						form_key = new_form_key
						val = new_val
					}
				} else if errors.Is(err, prefixtree.ErrPrefixAmbiguous) {
				} else {
					seq_end -= 1 // need
					// to backtrack one character since no match was found
					break
				}
			}
			if val != nil {
				metaform.Tokens = append(metaform.Tokens, val.(Token))
			}
			current_char = seq_end
		}
		doc.Metaforms = append(doc.Metaforms, &metaform)
		i++
	}

	return doc
}

//go:embed scripts/*
var scripts embed.FS

// Renders the handwritten output to the provided context
func Render(ctx *canvas.Context, options string, log *log.Logger) {
	var o Options
	// Unmarshal the JSON string into opts
	err := json.Unmarshal([]byte(options), &o)

	if err != nil {
		log.Println(err)
	}

	if o.Debug {
		log.Print("Options received: ")
		log.Printf("%+v\n", o)
	}

	log.SetFlags(0)		// remove timestamp

	// handwriting system definition
	var script *Script
	if o.Custom_script_svg_value != "" {
		script = load_script(o.Custom_script_svg_value)
	} else {
		file := fmt.Sprintf("scripts/%s", o.Builtin_script_name)
		script_file, err := scripts.ReadFile(file)
		if err != nil {
			log.Println("Error reading builtin script: ", err)
		}
		script = load_script(string(script_file))
	}

	// document to render
	d := Parse(o.Input_text, o.Language_code, script, log)
	if o.Debug {
		log.Print("Parsed document: ")
		log.Println(d)
	}

	// Used to write text
	var face *canvas.FontFace
	if o.Debug {
		fontNoto := canvas.NewFontFamily("noto")
		font_file, err := fonts.ReadFile("font/NotoSans-Light.ttf")
		if err != nil {
			log.Println("Error reading builtin font: ", err)
		}

		if err := fontNoto.LoadFont(font_file, 0, canvas.FontRegular); err != nil {
			panic(err)
		}
		face = fontNoto.Face(7.0, canvas.Black, canvas.FontBold, canvas.FontNormal)
	}

	width, height := ctx.Size()

	// Render each Metaform into m.Img with variable size and width
	for _, m := range d.Metaforms {
		m = m.renderForm(o, log)
	}

	pos := canvas.Point{X: o.Margin, Y: height - o.Margin}

	// Calculate which forms will fit on each line based on their width
	var lines [][]*Metaform
	lines = append(lines, []*Metaform{}) // initialize first line slice

	line_width := width - (o.Margin * 2)

	current_width := 0.0
	current_line := 0
	for _, m := range d.Metaforms {
		current_width += float64(m.Img.Bounds().Dx())
		if current_width <= line_width && m.Name != "\n" {
			lines[current_line] = append(lines[current_line], m)
			current_width += o.Space_between_metaforms
		} else {	// start a new line
			lines = append(lines, []*Metaform{m})
			current_width = float64(m.Img.Bounds().Dx()) + o.Space_between_metaforms // for next line
			current_line++
		}
	}

	// Redraw forms into the center of the line based on the tallest form in each line
	for _, line := range lines {
		max_height := 0.0
		for _, m := range line {
			max_height = math.Max(max_height, float64(m.Img.Bounds().Dy()))
		}
		for _, m := range line {
			m.centerForm(max_height, o, face)
		}
	}

	// Render each line
	for _, line := range lines {
		max_height := 0.0
		for _, m := range line {
			max_height = math.Max(max_height, float64(m.Img.Bounds().Dy()))
		}
		pos.Y -= max_height

		for _, m := range line {

			ctx.DrawImage(pos.X, pos.Y, m.Img, 1)
			pos.X += float64(m.Img.Bounds().Dx()) + o.Space_between_metaforms
		}
		pos.X = o.Margin
		pos.Y -= o.Space_between_lines
	}
}
