package main

import (
	"fmt"
	"mime/multipart"
	"net/http"
)


func httpRun(w http.ResponseWriter, r *http.Request) {
    sc := new_scrub(1, "amgen")
	httpInit(sc, sc.sr, r)
	sc.run()
}

func httpInit(sc *scrub, sr *scrub_req, r *http.Request) error {
    if err := r.ParseMultipartForm(1024*1024*100); err != nil {
        return err
    }
    httpParam(r.MultipartForm, &sr.auth, "auth", true)
    httpParam(r.MultipartForm, &sr.manu, "manu", true)
    httpParam(r.MultipartForm, &sr.sort, "sort", false)
    httpParam(r.MultipartForm, &sr.uniq, "uniq", false)

    httpFiles(r.MultipartForm, sc)
    if _, ok := sc.sr.files["rebates"];!ok {
        return fmt.Errorf("missing rebates file")
    }
    return nil
}

func httpFiles(form *multipart.Form, sc *scrub) {
    for name, list := range form.File {
        sf := &scrub_file{name: name}
        if len(list) > 0 {
            if mph, err := list[0].Open(); err == nil {
                sf.rdr = mph
            }
        }
        sc.sr.files[name] = sf
        httpParam(form, &sf.csep, fmt.Sprintf("%s_sep", name), false)
        httpParam(form, &sf.keys, fmt.Sprintf("%s_key", name), false)
    }
}
func httpParam(form *multipart.Form, dst *string, name string, req bool) {
    if vals, ok := form.Value[name];ok {
        if len(vals) > 0 {
            *dst = vals[0]
            return
        }
    }
    if req {
        if _, ok := form.Value["_missing_params"];!ok {
            form.Value["_missing_params"] = make([]string, 0, 1)
        }
        form.Value["_missing_params"] = append(form.Value["_missing_params"], name)
    }
}

