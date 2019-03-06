// embalses descarga, procesa y muestra niveles y status de los embalses de agua en Puerto Rico.
package main

import (
	"flag"
	"fmt"
	"log"
	_ "runtime/pprof"
	"strings"
	"time"
)

var verbose = flag.Bool("v", false, "Mostrar mensajes adicionales.")
var filter = flag.String("e", "", `Lista de embalses separados por coma: "Carite, La Plata, Toa Vaca"`)

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
	{"Carite","50039995",18.07524,-66.10683,544,542,539,537,536,8320, 0, 0, "", ""},
	{"Carraizo","50059000",18.32791,-66.01628,41.14,39.5,38.5,36.5,30,12000, 0, 0, "", ""},
	{"La Plata","50045000",18.343,-66.23607,51,43,39,38,31,26516, 0, 0, "", ""},
	{"Cidra","50047550",18.1969,-66.14072,403.05,401.05,400.05,399.05,398.05,4480, 0, 0, "", ""},
	{"Patillas","50093045",18.01774,-66.0185,67.07,66.16,64.33,60.52,59.45,9890, 0, 0, "", ""},
	{"Toa Vaca","50111210",18.10166,-66.48902,161,152,145,139,133,50650, 0, 0, "", ""},
	{"Rio Blanco","50076800",18.22389,-65.78142,28.75,26.5,24.25,22.5,18,3795, 0, 0, "", ""},
	{"Caonillas","50026140",18.27654,-66.65642,252,248,244,242,235,31730, 0, 0, "", ""},
	{"Fajardo","50071225",18.2969,-65.65858,52.5,48.3,43.4,37.5,26,4430, 0, 0, "", ""},
	{"Guajataca","50010800",18.39836,-66.9227,196,194,190,186,184,33340, 0, 0, "", ""},
	{"Cerrillos","50113950",18.07703,-66.57547,173.4,160,155.5,149.4,137.2,42600, 0, 0, "", ""},
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
	flag.Parse()
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
	if *filter != "" {
		fs := strings.Split(*filter, ",")
		onlyTheseSites = make(map[string]bool)
		for _, f := range fs {
			la := strings.ToLower(strings.Trim(f, " "))
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

	// Encabezado
	if *verbose {
		fmt.Println(header)
	}
	fmt.Printf("Actualizado: %s\n\n", today.Format("Jan 2, 2006 03:04:05"))

	// Nombres de columnas
	fmt.Printf("%-15s %-8s %-8s %-16s %-12s  %-8s %-9s %-8s %-8s %-8s\n", "Embalse", "Nivel", "Cambio", "Fecha medida", "Estatus", "Desborde", "Seguridad", "Observ", "Ajuste", "Control")

	// Output channel
	output := make(chan string, siteCount)

	// Benchmarking
	start := time.Now()

	// Process sites
	for _, site := range sites {
		// Filtrar embalses
		if len(onlyTheseSites) > 0 {
			ln := strings.ToLower(site.nombre)
			if !onlyTheseSites[ln] {
				continue
			}
		}

		// URL específico para info del embalse
		site.url = fmt.Sprintf(USGS_URL, site.id, yesterday.Format("2006.1.2"), today.Format("2006.1.2"))

		// Get site data concurrently
		go site.fetch(output)
	}

	// Recibir resultados del canal
	for i:=0; i<siteCount; i++ {
		fmt.Print(<-output)
	}

	// Mensajes finales
	if *verbose {
		fmt.Println(`
	Puede ver un embalse en específico, por ejemplo: embalses "La Plata"`)
		fmt.Printf("\nTiempo: %v\n", time.Since(start))
	}
}

func chk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
