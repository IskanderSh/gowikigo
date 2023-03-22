package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

type StringPage struct {
	Title string
	Body  string
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-z0-9]+)$")

func getPages() []StringPage {
	file, err := os.ReadFile("DataBase/titles.json")
	if err != nil {
		log.Fatal(err)
	}
	var titles []StringPage
	json.Unmarshal(file, &titles)
	return titles
}

func jsonSave(p *Page) {
	file, err := os.ReadFile("DataBase/titles.json")
	if err != nil {
		log.Fatal(file)
	}
	var titles []StringPage
	err = json.Unmarshal(file, &titles)
	if err != nil {
		log.Fatal(err)
	}
	flag := false
	for i, page := range titles {
		if p.Title == page.Title {
			titles[i] = StringPage{page.Title, string(p.Body)}
			flag = true
		}
	}
	if flag == false {
		titles = append(titles, StringPage{p.Title, string(p.Body)})
	}
	bytes, err := json.Marshal(titles)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("DataBase/titles.json", bytes, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Page) save() error {
	jsonSave(p)
	filename := "wikipages/" + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "wikipages/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func FrontPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("tmpl/front_page.html")
	titles := getPages()
	err := t.Execute(w, titles)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", FrontPage)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
