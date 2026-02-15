package service

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"strings"
	"vue-golang/internal/constants"
	"vue-golang/internal/storage"
)

type NormStorage interface {
	GetOrderMaterials(ctx context.Context, orderNum string, pos int) ([]*storage.KlaesMaterials, error)
	GetTemplateByCode(ctx context.Context, code string) (*storage.Template, error)
	GetDopInfoFromDemPrice(ctx context.Context, orderNum string) ([]*storage.DopInfoDemPrice, error)
}

type NormService struct {
	//storage *mysql.Storage
	storage NormStorage
}

func NewNormService(storage NormStorage) *NormService {
	return &NormService{storage: storage}
}

type Context struct {
	Type string

	HasImpost   bool
	ImpostCount float64

	StublinaCount float64

	//HasStvT   bool
	StvWindowCount float64
	StvTCount600   float64
	StvTCount400   float64

	MnogozapZamok float64
	StandZamok    float64

	PetliStand          float64
	PetliRolik          float64
	Petli3Section       float64
	HasPetliRDRH        bool
	PetliRDRH           float64
	HasPetliFural       bool
	PetliFural          float64
	PetliForNaveshCount float64

	PritvorKP40    float64
	HasPritvorKP40 bool

	TagCountWin float64

	//StvorkiWith3Petli float64
	// Добавишь больше признаков позже: тип профиля, площадь, кол-во камер и т.д.
}

//func (s *NormService) CalculateNorm(ctx context.Context, orderNum string, pos int, typeIzd string, templateCode string, itemCount int) ([]storage.Operation, Context, error) {
//	const op = "service.norm_service_rules.CalculateNorm"
//	// Получаем материалы
//	materials, err := s.storage.GetOrderMaterials(ctx, orderNum, pos)
//	if err != nil {
//		return nil, Context{}, fmt.Errorf("%s: ошибка получения всех материлов в сервисе %w", op, err)
//	}
//
//	//получаем примечание из dem_price
//	dopInfo, err := s.storage.GetDopInfoFromDemPrice(ctx, orderNum)
//	if err != nil {
//		return nil, Context{}, fmt.Errorf("%s: ошибка получения доп инфо из dem_price %w", op, err)
//	}
//
//	// Строим контекст
//	ctxData, err := BuildContext(materials, dopInfo, typeIzd)
//	if err != nil {
//		return nil, Context{}, fmt.Errorf("%s: ошибка построения контекста %w", op, err)
//	}
//
//	//fmt.Println("CTTTTXXXX", ctxData)
//
//	// Получаем шаблон (операции + правила)
//	template, err := s.storage.GetTemplateByCode(ctx, templateCode)
//	if err != nil {
//		return nil, Context{}, fmt.Errorf("%s: ошибка получения шаблона по коду в сервисе %w", op, err)
//	}
//
//	// Применяем правила
//	result := ApplyRules(template.Operations, template.Rules, ctxData, itemCount)
//
//	return result, ctxData, nil
//}

// TODO приколы с горутинами
func (s *NormService) CalculateNorm(ctx context.Context, orderNum string, pos int, typeIzd string, templateCode string, itemCount int) ([]storage.Operation, Context, error) {
	const op = "service.norm_service_rules.CalculateNorm"

	var (
		materials []*storage.KlaesMaterials
		template  *storage.Template
		dopInfo   []*storage.DopInfoDemPrice
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		materials, err = s.storage.GetOrderMaterials(gCtx, orderNum, pos)
		if err != nil {
			return fmt.Errorf("materials: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		template, err = s.storage.GetTemplateByCode(gCtx, templateCode)
		if err != nil {
			return fmt.Errorf("materials: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		dopInfo, err = s.storage.GetDopInfoFromDemPrice(gCtx, orderNum)
		if err != nil {
			return fmt.Errorf("materials: %w", err)
		}
		return nil
	})

	err := g.Wait()
	if err != nil {
		return nil, Context{}, err
	}

	buildContext, err := BuildContext(materials, dopInfo, typeIzd)
	if err != nil {
		return nil, Context{}, err
	}

	result := ApplyRules(template.Operations, template.Rules, buildContext, itemCount)

	return result, buildContext, nil
}

func BuildContextGlyhar(materials []*storage.KlaesMaterials) Context {
	ctx := Context{Type: "glyhar"}

	for _, m := range materials {

		name := strings.TrimSpace(m.NameMat)
		//if name == "импост" || name == "Профиль импостный" || name == "стойка-импост" || name == "импост в дверь" {
		//	ctx.HasImpost = true
		//	ctx.ImpostCount += m.Count
		//}

		if constants.ImpostCount[name] {
			ctx.HasImpost = true
			ctx.ImpostCount += m.Count
		}

		if constants.StvTCount600[name] && m.Width <= 615 {
			ctx.StvTCount600 += m.Count
		}

		if constants.StvTCount400[name] && m.Width <= 400 {
			ctx.StvTCount400 += m.Count
		}
	}

	//log.Printf("Смотрим материалы: HasImpost=%v, ImpostCount=%f", ctx.HasImpost, ctx.ImpostCount)

	return ctx
}

func BuildContextWindow(materials []*storage.KlaesMaterials) Context {
	ctx := Context{Type: "window"}

	for _, m := range materials {
		name := strings.TrimSpace(m.NameMat)

		if constants.ImpostCount[name] || constants.ShtylpCount[name] {
			ctx.HasImpost = true
			ctx.ImpostCount += m.Count
		}

		if constants.StvWindow[name] {
			ctx.StvWindowCount += m.Count
		}

		if constants.StvTCount600[name] && m.Width <= 615 {
			ctx.StvTCount600 += m.Count
		}

		if constants.StvTCount400[name] && m.Width <= 400 {
			ctx.StvTCount400 += m.Count
		}

		if constants.TagCountWin[name] {
			ctx.TagCountWin += m.Count
		}

	}

	log.Printf("Смотрим материалы: HasImpost=%v, ImpostCount=%f, StvCount=%f, TagCount=%f", ctx.HasImpost, ctx.ImpostCount, ctx.StvWindowCount, ctx.TagCountWin)

	return ctx
}

func BuildContextDoor(materials []*storage.KlaesMaterials, dopInfo []*storage.DopInfoDemPrice) Context {
	ctx := Context{Type: "door"}

	for _, m := range materials {

		//log.Printf("MATER", m)
		name := strings.TrimSpace(m.NameMat)
		//if name == "импост" || name == "профиль импостный" || name == "стойка-импост" || name == "импост в дверь" {
		//	ctx.HasImpost = true
		//	ctx.ImpostCount += m.Count
		//}

		if constants.ImpostCount[name] {
			ctx.HasImpost = true
			ctx.ImpostCount += m.Count
		}

		//if name == "накладка на цилиндр stublina" || name == "накладка на цилиндр stublina (под покраску)" {
		//	ctx.StublinaCount += m.Count
		//}

		if constants.StublinaCount[name] {
			ctx.StublinaCount += m.Count
		}

		//if (name == "створка т-образная" || name == "створка-коробка" || name == "створка т - образ.") && m.Width < 615 {
		//	//ctx.HasStvT = true
		//	ctx.StvTCount600 += m.Count
		//	//log.Printf("MATERILAs", m.Width)
		//}

		if constants.StvTCount600[name] && m.Width <= 615 {
			ctx.StvTCount600 += m.Count
		}

		//if (name == "створка т-образная" || name == "створка-коробка" || name == "створка т - образ.") && m.Width < 400 {
		//	//ctx.HasStvT = true
		//	ctx.StvTCount400 += m.Count
		//	//log.Printf("MATERILAs", m.Width)
		//}

		if constants.StvTCount400[name] && m.Width <= 400 {
			ctx.StvTCount400 += m.Count
		}

		//TODO Притвор КП40
		//if name == "притвор кп40" {
		//	ctx.PritvorKP40 += m.Count
		//	ctx.HasPritvorKP40 = true
		//}

		if constants.PritvorKP40[name] {
			ctx.PritvorKP40 += m.Count
			ctx.HasPritvorKP40 = true
		}

		//TODO ПЕТЛИ

		if constants.PetliStand[name] {
			ctx.PetliStand += m.Count
			ctx.PetliForNaveshCount += m.Count
		}

		if constants.PetliRolik[name] {
			ctx.PetliRolik += m.Count
			ctx.PetliForNaveshCount += m.Count
		}

		if constants.Petli3Section[name] {
			ctx.Petli3Section += m.Count
		}

		if constants.PetliFural[name] {
			ctx.PetliFural += m.Count
			ctx.HasPetliFural = true
			ctx.PetliForNaveshCount += m.Count
		}

		if constants.PetliRDRH[name] {
			ctx.PetliRDRH += m.Count
			ctx.HasPetliRDRH = true
		}

		//zamokk := make(map[int]string)
		//
		////TODO замок
		//if name == "многозапорный замок stublina с управлением от ручки" {
		//	ctx.MnogozapZamok += m.Count
		//	ctx.StandZamok += m.Count
		//}

		if constants.MnogozapZamok[name] {
			ctx.MnogozapZamok += m.Count
			ctx.StandZamok += m.Count
		}

		if constants.StandZamok[name] {
			ctx.StandZamok += m.Count
		}

		// TODO функция для доп замка в dem_price потом раскометировать
		//for _, price := range dopInfo {
		//	addCount, err := strconv.Atoi(price.Position)
		//	if err != nil {
		//		log.Printf("Ошибка преобразования строки в число %s", err)
		//		continue
		//	}
		//
		//	if m.ArticulMat == price.ArticulMat && m.Position == addCount {
		//		ctx.StandZamok += price.Count
		//		log.Printf("Найден доп. замок: артикул=%s, позиция=%d, доп. кол-во=%.0f",
		//			m.ArticulMat, m.Position, price.Count)
		//		break
		//	}
		//}
	}

	//log.Printf("Смотрим материалы: HasImpost=%v, ImpostCount=%f, StublinaCount=%v, StvTCount600=%f, PetliRDRH=%v, MnogoazapoR=%f, PetliRolik=%v, PetliStand=%v,Petli#Section=%v, Pritvor=%v, PetliFural=%v, StandZamok=%v",
	//	ctx.HasImpost, ctx.ImpostCount, ctx.StublinaCount, ctx.StvTCount600, ctx.HasPetliRDRH, ctx.MnogozapZamok, ctx.PetliRolik, ctx.PetliStand, ctx.Petli3Section, ctx.PritvorKP40, ctx.HasPetliFural, ctx.StandZamok)

	log.Printf("Смотрим материалы: PetliStand=%v, PetliRolik=%v, Petli3Section=%v, PetliFural=%v, PetliRDRH=%v, PetliForNaveshCount=%v", ctx.PetliStand, ctx.PetliRolik, ctx.Petli3Section, ctx.PetliFural, ctx.PetliRDRH, ctx.PetliForNaveshCount)

	return ctx
}

func BuildContext(materials []*storage.KlaesMaterials, dopInfo []*storage.DopInfoDemPrice, typeIzd string) (Context, error) {
	switch typeIzd {
	case "glyhar":
		return BuildContextGlyhar(materials), nil
	case "window":
		return BuildContextWindow(materials), nil
	case "door":
		return BuildContextDoor(materials, dopInfo), nil
	default:
		return Context{}, fmt.Errorf("неизвестный тип изделия: %s", typeIzd)
	}
}

func ApplyRules(operations []storage.Operation, rules []storage.Rule, ctx Context, itemCount int) []storage.Operation {
	result := make([]storage.Operation, len(operations))
	copy(result, operations)

	//log.Printf("Загружено правил: %d", len(rules))
	//for i, r := range rules {
	//	log.Printf("Правило %d: op=%s, cond=%v", i, r.Operation, r.Condition)
	//}
	//itemCount := 2

	for i := range result {
		if result[i].Group != "ign" {
			result[i].Value *= float64(itemCount)
			result[i].Minutes *= float64(itemCount)
			result[i].Count *= float64(itemCount)
		}
	}

	for i := range result {
		for _, rule := range rules {
			if rule.Operation != result[i].Name {
				continue
			}

			if !MatchesCondition(rule.Condition, ctx) {
				continue
			}

			switch rule.Mode {
			case "set":
				result[i].Value = rule.SetValue
				result[i].Minutes = rule.SetMinutes
			case "multiplied":
				count := getCountMaterials(rule.UnitField, ctx, itemCount)
				//log.Printf("GGGGGGGG", count)
				result[i].Value = rule.ValuePerUnit * count
				result[i].Minutes = rule.MinutesPerUnit * count
				result[i].Count = count
			case "additive":
				result[i].Value += rule.ValuePerUnit
				result[i].Minutes += rule.MinutesPerUnit
			case "additivePlusMultiplied":
				count := getCountMaterials(rule.UnitField, ctx, itemCount)
				result[i].Value += rule.ValuePerUnit * count
				result[i].Minutes += rule.MinutesPerUnit * count
				result[i].Count += count
			case "minus":
				result[i].Value -= rule.ValuePerUnit
				result[i].Minutes -= rule.MinutesPerUnit
			default:
				// По умолчанию — просто замена
				//result[i].Value = rule.SetValue
				//result[i].Minutes = rule.SetMinutes
			}
			//break // применили первое подходящее правило
		}
	}

	//log.Printf("RULES", result)

	return result
}

func getCountMaterials(field string, ctx Context, itemCount int) float64 {
	switch field {
	case "HasImpostCount":
		return ctx.ImpostCount
	case "StvTCount600":
		return ctx.StvTCount600
	case "StvTCount400":
		return ctx.StvTCount400
	case "MnogozapZamok":
		return ctx.MnogozapZamok
	case "StandZamok":
		return ctx.StandZamok
	case "PetliRolik":
		return ctx.PetliRolik
	case "PetliStand":
		return ctx.PetliStand
	case "Petli3Section":
		return ctx.Petli3Section
	case "StublinaCount":
		return ctx.StublinaCount
	case "PritvorKP40":
		return ctx.PritvorKP40
	case "PetliFural":
		return ctx.PetliFural
	case "PetliRDRH":
		return ctx.PetliRDRH
	case "HasPritvorKP40":
		return float64(itemCount)
	case "StvWindowCount":
		return ctx.StvWindowCount
	case "TagCountWin":
		return ctx.TagCountWin
	case "PetliForNaveshCount":
		return ctx.PetliForNaveshCount
	case "ItemCountForRDRH":
		if ctx.HasPetliRDRH {
			return float64(itemCount)
		}
		return 0
	//case "StvorkiWith3Petli":
	//return ctx.StvorkiWith3Petli
	default:
		return 0
	}
}

func MatchesCondition(condition map[string]interface{}, ctx Context) bool {
	//log.Printf("CONDITION", condition)
	for key, expected := range condition {
		if !fieldMatches(key, expected, ctx) {
			return false
		}
	}
	return true
}

func fieldMatches(key string, expected interface{}, ctx Context) bool {
	//log.Printf("FIELD MATCHES", key, expected, ctx)
	switch key {
	case "HasImpost":
		if val, ok := expected.(bool); ok {
			return ctx.HasImpost == val
		}
	case "StublinaCount":
		return compareFloatField(ctx.StublinaCount, expected)
	case "StvTCount600":
		return compareFloatField(ctx.StvTCount600, expected)
	case "StvTCount400":
		return compareFloatField(ctx.StvTCount400, expected)
	case "HasPetliRDRH":
		if val, ok := expected.(bool); ok {
			return ctx.HasPetliRDRH == val
		}
	case "HasPritvorKP40":
		if val, ok := expected.(bool); ok {
			return ctx.HasPritvorKP40 == val
		}
	case "MnogozapZamok":
		return compareFloatField(ctx.MnogozapZamok, expected)
	case "StandZamok":
		return compareFloatField(ctx.StandZamok, expected)
	case "PetliRolik":
		return compareFloatField(ctx.PetliRolik, expected)
	case "PetliStand":
		return compareFloatField(ctx.PetliStand, expected)
	case "Petli3Section":
		return compareFloatField(ctx.Petli3Section, expected)
	case "PritvorKP40":
		return compareFloatField(ctx.PritvorKP40, expected)
	case "HasPetliFural":
		if val, ok := expected.(bool); ok {
			return ctx.HasPetliFural == val
		}
	case "PetliFural":
		return compareFloatField(ctx.PetliFural, expected)
	case "PetliRDRH":
		return compareFloatField(ctx.PetliRDRH, expected)
	case "StvWindowCount":
		return compareFloatField(ctx.StvWindowCount, expected)
	case "TagCountWin":
		return compareFloatField(ctx.TagCountWin, expected)
	case "PetliForNaveshCount":
		return compareFloatField(ctx.PetliForNaveshCount, expected)
	//case "StvorkiWith3Petli":
	//return compareFloatField(ctx.StvorkiWith3Petli, expected)
	//case "ImpostCount":
	// Поддержка: {"min": 2} или просто 2
	//return compareIntField(ctx.ImpostCount, expected)
	default:
		return false
	}
	return false
}

func compareFloatField(actual float64, expected interface{}) bool {
	// Случай 1: ожидается конкретное число
	if val, ok := expected.(float64); ok {
		return actual == val
	}
	// Случай 2: ожидается объект { "min": 1 } или { "min": 1, "max": 5 }
	if obj, ok := expected.(map[string]interface{}); ok {
		if minVal, hasMin := obj["min"].(float64); hasMin {
			if actual < minVal {
				return false
			}
		}
		if maxVal, hasMax := obj["max"].(float64); hasMax {
			if actual > maxVal {
				return false
			}
		}
		return true
	}
	return false
}
