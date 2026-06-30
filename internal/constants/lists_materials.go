package constants

var (
	// TODO замки
	MnogozapZamok = map[string]bool{
		"Многозапорный замок Stublina с управлением от ручки":   true,
		"Многозапорный замок KFV AS4350 с управлением от ручки": true,
		"Многозапорный замок KFV AS2750 с управлением от ключа": true,
		"Многозапорный замок Fuhr D30 с управлением от ключа":   true,
	}

	StandZamok = map[string]bool{
		"Замок Elementis 1153 (D30) (под нажимной гарнитур)": true,
		"Замок MACO G-TS 57819(232011)":                      true,
		"Замок KALE 153 D30мм E85 N8 (с защёлкой)":           true,
		"Замок Elementis 1155 (D30) (для бугельных ручек)":   true,
	}

	// TODO импост
	ImpostCount = map[string]bool{
		"Стойка-импост":               true,
		"Импост":                      true,
		"Профиль импостный":           true,
		"Импост в дверь":              true,
		"Ригель облег. двухпол. КП40": true,
		"Стойка-импост 64мм":          true,
	}

	ShtylpCount = map[string]bool{
		"Штульп": true,
	}

	// TODO накладки стаблина
	StublinaCount = map[string]bool{
		"Накладка на цилиндр Stublina":                true,
		"Накладка на цилиндр Stublina (под покраску)": true,
	}

	//TODO створки
	StvWindow = map[string]bool{
		"Створка Т-образная":                      true,
		"Створка-коробка":                         true,
		"Створка Т - образ.":                      true,
		"Створка оконная усиленная прямоугольная": true,
		"Створка оконная":                         true,
		"Створка оконная усиленная":               true,
		"Створка": true,
	}

	StvTCount600 = map[string]bool{
		"Створка Т-образная":                      true,
		"Створка-коробка":                         true,
		"Створка Т - образ.":                      true,
		"Створка оконная":                         true,
		"Створка оконная усиленная прямоугольная": true,
		"Створка оконная усиленная":               true,
		"Створка": true,
	}

	StvTCount400 = map[string]bool{
		"Створка Т-образная":                      true,
		"Створка-коробка":                         true,
		"Створка Т - образ.":                      true,
		"Створка оконная":                         true,
		"Створка оконная усиленная прямоугольная": true,
		"Створка оконная усиленная":               true,
		"Створка": true,
	}

	//TODO тяги
	TagCountWin = map[string]bool{
		"Фурнитурная тяга":           true,
		"03524590N Фурнитурная тяга": true,
	}

	//TODO притвор
	PritvorKP40 = map[string]bool{
		"Притвор КП40": true,
	}

	// TODO петли
	PetliStand = map[string]bool{
		"Петля двухсекционная 67мм": true,
		"Петля дверная 2-част.":     true,
	}

	PetliRolik = map[string]bool{
		"Петля роликовая для КП45": true,
	}

	Petli3Section = map[string]bool{
		"Петля дверная трехсекционная с удлиненной базой":      true,
		"Петля дверная трехсекционная с удлиненной базой  :::": true,
		"Петля Фурал дверная 3-част. с подшипником":            true,
	}

	PetliFural = map[string]bool{
		"Петля Фурал дверная 2-част. с подшипником": true,
	}

	PetliRDRH = map[string]bool{
		"Петля роликовая RDRH": true,
	}

	//TODO рамы глухарей(пока для операции доп напиловки/сборки)
	RamRigelGl = map[string]bool{
		"Стойка ригель. глухарей": true,
		"Рама": true,
	}
)
