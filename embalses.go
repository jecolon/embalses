/*  Traducción en Go del script en Python por Edwood Ocasio: https://github.com/eocasio/embalses
Inspirado por http://mate.uprh.edu/embalsespr/
8/31/2015
3/05/2019 - Refactor

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston,
MA 02110-1301, USA.

Copyright 2015 Edwood Ocasio <edwood.ocasio@gmail.com>
Ported to Go by Jose Colon <jec.rod@gmail.com>
*/

//  embalses muestra cambio en nivel en ultimas 24 horas (aproximadamente)
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	_ "runtime/pprof"
	"strconv"
	"strings"
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
}

var header = `
Muestra cambio en nivel en ultimas 24 horas (aproximadamente)	
Versión Go de Edwood Ocasio's: https://github.com/eocasio/embalses
Inspirado por http://mate.uprh.edu/embalsespr/
Copyright 2019 Jose E. Colon <jec.rod@gmail.com>

Datos provistos por el grupo CienciaDatosPR del Departamento de 
Matematicas de la Universidad de Puerto Rico en Humacao	
Estos datos están sujetos a revisión por el USGS y no deben ser	
tomados como oficiales o libres de errores de medición.	
`

var sites = []*embalse{
	{"Carite","50039995",18.07524,-66.10683,544,542,539,537,536,8320, 0, 0, ""},
	{"Carraizo","50059000",18.32791,-66.01628,41.14,39.5,38.5,36.5,30,12000, 0, 0, ""},
	{"La Plata","50045000",18.343,-66.23607,51,43,39,38,31,26516, 0, 0, ""},
	{"Cidra","50047550",18.1969,-66.14072,403.05,401.05,400.05,399.05,398.05,4480, 0, 0, ""},
	{"Patillas","50093045",18.01774,-66.0185,67.07,66.16,64.33,60.52,59.45,9890, 0, 0, ""},
	{"Toa Vaca","50111210",18.10166,-66.48902,161,152,145,139,133,50650, 0, 0, ""},
	{"Rio Blanco","50076800",18.22389,-65.78142,28.75,26.5,24.25,22.5,18,3795, 0, 0, ""},
	{"Caonillas","50026140",18.27654,-66.65642,252,248,244,242,235,31730, 0, 0, ""},
	{"Fajardo","50071225",18.2969,-65.65858,52.5,48.3,43.4,37.5,26,4430, 0, 0, ""},
	{"Guajataca","50010800",18.39836,-66.9227,196,194,190,186,184,33340, 0, 0, ""},
	{"Cerrillos","50113950",18.07703,-66.57547,173.4,160,155.5,149.4,137.2,42600, 0, 0, ""},
}

func isValidSite(s string) bool {
	valid := false
	for _, site := range sites {
		if strings.ToLower(s) == strings.ToLower(site.nombre) {
			valid = true
		}
	}
	return valid
}

func main() {
	/*
		f, err := os.Create("cpu.out")
		chk(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/

	// Lista de embalses específicos por nombre
	// Deje lista vacía para incluir todos los embalses.
	siteCount := len(sites)
	var onlyTheseSites map[string]bool
	if len(os.Args) > 1 {
		onlyTheseSites = make(map[string]bool)
		for _, a := range os.Args[1:] {
			la := strings.ToLower(a)
			if isValidSite(la) {
				onlyTheseSites[la] = true
			}
		}
		if len(onlyTheseSites) > 0 {
			siteCount = len(onlyTheseSites)
		}
	}

	// Fechas de interés
	today := time.Now()
	yesterday := today.Add(-24 * time.Hour)

	// Plantilla URL para obtener datos del USGS
	// ver http://waterdata.usgs.gov/nwis?automated_retrieval_info
	USGS_URL := "http://nwis.waterdata.usgs.gov/pr/nwis/uv/?cb_62616=on&format=rdb&site_no=%s&begin_date=%s&end_date=%s"

	fmt.Println(header)
	fmt.Printf("Actualizado: %s\n\n", today.Format("Jan 2, 2006 03:04:05"))

	// Nombre columnas
	fmt.Printf("%-15s %-8s %-8s %-16s %-12s  %-8s %-9s %-8s %-8s %-8s\n", "Embalse", "Nivel", "Cambio", "Fecha medida", "Estatus", "Desborde", "Seguridad", "Observ", "Ajuste", "Control")

	// Output channel
	output := make(chan string, siteCount)

	// Benchmarking
	start := time.Now()

	// Process sites
	for _, site := range sites {
		// ¿Existe una lista específica de embalses?
		if len(onlyTheseSites) > 0 {
			ln := strings.ToLower(site.nombre)
			if !onlyTheseSites[ln] {
				continue
			}
		}

		// URL específico para info del embalse
		urlFinal := fmt.Sprintf(USGS_URL, site.id, yesterday.Format("2006.1.2"), today.Format("2006.1.2"))

		// Get site data concurrently
		go getSiteData(site, urlFinal, output)
	}

	for i:=0; i<siteCount; i++ {
		fmt.Print(<-output)
	}

	fmt.Println(`
Puede ver un embalse en específico, por ejemplo: embalses "La Plata"`)
	fmt.Printf("\nTiempo: %v\n", time.Since(start))
}

func getSiteData(site *embalse, urlFinal string, output chan<- string) {
	// Set network call timeout.
	httpClient := &http.Client{
		Timeout: 1 * time.Minute,
	}

	// Buscar datos
	r, err := httpClient.Get(urlFinal)
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

func chk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
