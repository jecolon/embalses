/*  Traducción en Go del script en Python por Edwood Ocasio: https://github.com/eocasio/embalses
Inspirado por http://mate.uprh.edu/embalsespr/
8/31/2015

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

//  embalsespr muestra cambio en nivel en ultimas 24 horas (aproximadamente)
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "os"
	_ "runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

var header = `########################################################################
#  embalsespr
#  Muestra cambio en nivel en ultimas 24 horas (aproximadamente)
#  Inspirado por http://mate.uprh.edu/embalsespr/
#  8/27/2015
#  Copyright 2015 Edwood Ocasio <edwood.ocasio@gmail.com>
########################################################################
# Datos provistos por el grupo CienciaDatosPR del Departamento de 
# Matematicas de la Universidad de Puerto Rico en Humacao
# Inspirado por http://mate.uprh.edu/embalsespr/
#
# Estos datos están sujetos a revisión por el USGS y no deben ser
# tomados como oficiales o libres de errores de medición.
########################################################################
`

/*
 Datos provistos por el grupo CienciaDatosPR del Departamento de
 Matemáticas de la Universidad de Puerto Rico en Humacao
 https://raw.githubusercontent.com/mecobi/EmbalsesPR/master/embalses.csv
*/
var sitesRawData = `Carite,50039995,18.07524,-66.10683,544,542,539,537,536,8320
Carraizo,50059000,18.32791,-66.01628,41.14,39.5,38.5,36.5,30,12000
La Plata,50045000,18.343,-66.23607,51,43,39,38,31,26516
Cidra,50047550,18.1969,-66.14072,403.05,401.05,400.05,399.05,398.05,4480
Patillas,50093045,18.01774,-66.0185,67.07,66.16,64.33,60.52,59.45,9890
Toa Vaca,50111210,18.10166,-66.48902,161,152,145,139,133,50650
Rio Blanco,50076800,18.22389,-65.78142,28.75,26.5,24.25,22.5,18,3795
Caonillas,50026140,18.27654,-66.65642,252,248,244,242,235,31730
Fajardo,50071225,18.2969,-65.65858,52.5,48.3,43.4,37.5,26,4430
Guajataca,50010800,18.39836,-66.9227,196,194,190,186,184,33340
Cerrillos,50113950,18.07703,-66.57547,173.4,160,155.5,149.4,137.2,42600`

func main() {
	/*
		f, err := os.Create("cpu.out")
		chk(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/

	sitesDataLines := strings.Split(sitesRawData, "\n")
	sitesData := make([][]string, len(sitesDataLines))
	for i, l := range sitesDataLines {
		sitesData[i] = strings.Split(l, ",")
	}

	// Lista de embalses específicos por nombre
	// Deje lista vacía para incluir todos los embalses.
	// e.g.
	// onlyTheseSites := map[string]bool{
	//   "Carite": true,
	// }
	onlyTheseSites := make(map[string]bool)

	// Fechas de interés
	today := time.Now()
	yesterday := today.Add(-24 * time.Hour)

	// Plantilla URL para obtener datos del USGS
	// ver http://waterdata.usgs.gov/nwis?automated_retrieval_info
	USGS_URL := "http://nwis.waterdata.usgs.gov/pr/nwis/uv/?cb_62616=on&format=rdb&site_no=%s&begin_date=%s&end_date=%s"

	fmt.Println(header)
	fmt.Printf("Actualizado: %s\n", today.String())
	fmt.Println("Datos embalses en USGS...\n")

	// Nombre columnas
	fmt.Printf("%-15s %-8s %-8s %-16s %-12s  %-8s %-9s %-8s %-8s %-8s\n", "Embalse", "Nivel", "Cambio", "Fecha medida", "Estatus", "Desborde", "Seguridad", "Observ", "Ajuste", "Control")

	// Concurrency
	output := make(chan string, 11)
	gophers := 0

	for _, site := range sitesData {
		// Site name
		siteName := site[0]

		// ¿Existe una lista específica de embalses?
		if len(onlyTheseSites) > 0 {
			if _, ok := onlyTheseSites[siteName]; !ok {
				continue
			}
		}

		// ID del embalse
		siteID := site[1]

		// URL específico para info del embalse
		urlFinal := fmt.Sprintf(USGS_URL, siteID, yesterday.Format("2006.1.2"), today.Format("2006.1.2"))

		// Get data concurrently
		wg.Add(1)
		gophers++
		go getSiteData(site, urlFinal, output)
	}

	var buf bytes.Buffer
	for i := 0; i < gophers; i++ {
		buf.WriteString(<-output)
	}
	fmt.Print(buf.String())
	wg.Wait()
}

func getSiteData(site []string, urlFinal string, output chan<- string) {
	// Buscar datos
	r, err := http.Get(urlFinal)
	chk(err)

	// Convertirlos en slice
	b, err := ioutil.ReadAll(r.Body)
	chk(err)
	defer r.Body.Close()

	// Extraer todas las filas de datos ignorando comentarios
	levelLines := strings.Split(string(b), "\n")
	rows := make([][]string, 0)
	for _, l := range levelLines {
		if len(l) == 0 || l[0] == '#' {
			continue
		}
		fields := strings.Split(l, "\t")
		if fields[0] == "USGS" {
			rows = append(rows, fields)
		}
	}

	// Obtener nivel del embalse hace 24 horas (aproximadamente)
	// Cada fila representa un intervalo de 15 minutos.
	// La fila 96 intervalos atrás contiene ese valor (aproximadamente).
	firstLevel, err := strconv.ParseFloat(rows[0][4], 64)
	chk(err)

	// Última lectura del archivo
	lastRow := rows[len(rows)-1]
	measurementDate := lastRow[2]
	lastLevel, err := strconv.ParseFloat(lastRow[4], 64)
	chk(err)

	diffLevels := lastLevel - firstLevel

	// Determinar estatus del embalse
	desborde, err := strconv.ParseFloat(site[4], 64)
	chk(err)
	seguridad, err := strconv.ParseFloat(site[5], 64)
	chk(err)
	observacion, err := strconv.ParseFloat(site[6], 64)
	chk(err)
	ajuste, err := strconv.ParseFloat(site[7], 64)
	chk(err)
	control, err := strconv.ParseFloat(site[8], 64)
	chk(err)

	// Status switch
	status := "FUERA SERVICIO"
	switch {
	case lastLevel >= desborde:
		status = "DESBORDE"
	case lastLevel >= seguridad:
		status = "SEGURIDAD"
	case lastLevel >= observacion:
		status = "OBSERVACION"
	case lastLevel >= ajuste:
		status = "AJUSTE"
	case lastLevel >= control:
		status = "CONTROL"
	}

	// Mostrar información resumida del embalse:
	// nombre, última lectura en metros, cambio en últimas 24 horas, fecha última lectura, estatus, estatus y sus niveles de referencia
	output <- fmt.Sprintf("%-15s %-8.2f [%-5.2fm] %-16s %-12s  %-8.2f %-9.2f %-8.2f %-8.2f %-8.2f\n", site[0], lastLevel, diffLevels, measurementDate, status, desborde, seguridad, observacion, ajuste, control)
	wg.Done()
}

func chk(err error) {
	if err != nil {
		log.Fatalf("ERROR: %s", err.Error())
	}
}
