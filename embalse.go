package main

import(
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type embalse struct {
	nombre      string
	id          string
	lat         float64
	lng         float64
	desborde    float64
	seguridad   float64
	observacion float64
	ajuste      float64
	control     float64
	capacidad   float64
	nivelPrevio float64
	nivelActual float64
	status      string
	url         string
}

func (site *embalse) fetch(output chan<- string) {
	// Set network call timeout.
	httpClient := &http.Client{
		Timeout: 1 * time.Minute,
	}

	// Buscar datos
	r, err := httpClient.Get(site.url)
	if err != nil {
		output <- fmt.Sprintf("error descargando datos para %s: %s\n", site.nombre, err)
		return
	}
	defer r.Body.Close()

	// Convertirlos en slice
	levelReader := csv.NewReader(r.Body)
	levelReader.Comma = '\t'
	levelReader.Comment = '#'
	allRows, err := levelReader.ReadAll()
	if err != nil {
		output <- fmt.Sprintf("error procesando datos para %s: %s\n", site.nombre, err)
		return
	}

	rows := make([][]string, 0)
	for _, r := range allRows {
		if r[0] == "USGS" {
			rows = append(rows, r)
		}
	}

	// Obtener nivel del embalse hace 24 horas (aproximadamente)
	// Cada fila representa un intervalo de 15 minutos.
	// La fila 96 intervalos atrás contiene ese valor (aproximadamente).
	site.nivelPrevio, err = strconv.ParseFloat(rows[0][4], 64)
	if err != nil {
		output <- fmt.Sprintf("error procesando datos para %s: %s\n", site.nombre, err)
		return
	}

	// Última lectura del archivo
	lastRow := rows[len(rows)-1]
	measurementDate := lastRow[2]
	site.nivelActual, err = strconv.ParseFloat(lastRow[4], 64)
	if err != nil {
		output <- fmt.Sprintf("error procesando datos para %s: %s\n", site.nombre, err)
		return
	}

	diffLevels := site.nivelActual - site.nivelPrevio

	// Determinar estatus del embalse
	site.status = "FUERA SERVICIO"
	switch {
	case site.nivelActual >= site.desborde:
		site.status = "DESBORDE"
	case site.nivelActual >= site.seguridad:
		site.status = "SEGURIDAD"
	case site.nivelActual >= site.observacion:
		site.status = "OBSERVACION"
	case site.nivelActual >= site.ajuste:
		site.status = "AJUSTE"
	case site.nivelActual >= site.control:
		site.status = "CONTROL"
	}

	// Mostrar información resumida del embalse:
	output <- fmt.Sprintf("%-15s %-8.2f [%-5.2fm] %-16s %-12s  %-8.2f %-9.2f %-8.2f %-8.2f %-8.2f\n", site.nombre, site.nivelActual, diffLevels, measurementDate, site.status, site.desborde, site.seguridad, site.observacion, site.ajuste, site.control)
}

