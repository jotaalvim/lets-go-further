package main

import (
	"net/http"
)

func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {

	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	//js := `{ "status": "available", "environment": %q, "version": %q }`
	//js = fmt.Sprintf(js, app.config.env, version)

	err := app.writeJSON(w, http.StatusOK, data, nil)

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, " encontrei um erro", http.StatusInternalServerError)
	}

}
