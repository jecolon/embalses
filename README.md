# embalses
Go port of @eocasio 's Python script to fetch Puerto Rico water basin surface level data.

# Building
Go 1.5+ recommended.
This assumes your $GOPATH/bin is in your $PATH. If not, prefix $GOPATH/bin/ when running binary.

```bash
$ go get github.com/jecolon/embalses
$ embalses
```

# Usage
```bash
$ embalses -h
Usage of embalses:
  -e string
    	Lista de embalses separados por coma: "Carite, La Plata, Toa Vaca"
  -v	Mostrar mensajes adicionales.

$ embalses
Actualizado: Mar 6, 2019 11:13:28

Embalse         Nivel    Cambio   Fecha medida     Estatus       Desborde Seguridad Observ   Ajuste   Control
Patillas        64.66    [0.04 m] 2019-03-06 10:45 OBSERVACION   67.07    66.16     64.33    60.52    59.45
Guajataca       188.00   [-0.03m] 2019-03-06 10:45 AJUSTE        196.00   194.00    190.00   186.00   184.00
Toa Vaca        149.61   [-0.08m] 2019-03-06 10:15 OBSERVACION   161.00   152.00    145.00   139.00   133.00
Carraizo        40.80    [0.00 m] 2019-03-06 10:45 SEGURIDAD     41.14    39.50     38.50    36.50    30.00
Carite          542.83   [0.02 m] 2019-03-06 10:15 SEGURIDAD     544.00   542.00    539.00   537.00   536.00
Fajardo         52.47    [0.01 m] 2019-03-06 10:46 SEGURIDAD     52.50    48.30     43.40    37.50    26.00
Rio Blanco      27.50    [-0.01m] 2019-03-06 10:30 SEGURIDAD     28.75    26.50     24.25    22.50    18.00
Cerrillos       171.10   [-0.05m] 2019-03-06 10:45 SEGURIDAD     173.40   160.00    155.50   149.40   137.20
Cidra           401.86   [-0.02m] 2019-03-06 10:45 SEGURIDAD     403.05   401.05    400.05   399.05   398.05
Caonillas       250.37   [-0.45m] 2019-03-06 11:00 SEGURIDAD     252.00   248.00    244.00   242.00   235.00
La Plata        49.62    [-0.07m] 2019-03-06 10:45 SEGURIDAD     51.00    43.00     39.00    38.00    31.00

$ embalses -e "Carite, Toa VAca"
Actualizado: Mar 6, 2019 11:23:28

Embalse         Nivel    Cambio   Fecha medida     Estatus       Desborde Seguridad Observ   Ajuste   Control
Carite          542.83   [0.01 m] 2019-03-06 11:15 SEGURIDAD     544.00   542.00    539.00   537.00   536.00
Toa Vaca        149.61   [-0.08m] 2019-03-06 10:15 OBSERVACION   161.00   152.00    145.00   139.00   133.00
```
