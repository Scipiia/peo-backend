package mysql

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

type TestOrderFixture struct {
	OrderNum  string
	Creator   int
	Customer  string
	DopInfo   string
	MsNote    string
	Year      int
	Month     int
	Positions []TestPosition
	Images    []TestImage
	Plans     []TestPlan
}

type TestPosition struct {
	Position     string
	NamePosition string
	Count        string
}

type TestImage struct {
	OrderName string
	OrderPos  string
	Image     string
}

type TestPlan struct {
	IDOrder int64
	X       string // position
	Color   string
	Sqr     float64
}

func createTestOrderDem(t *testing.T, fixture TestOrderFixture) int64 {

	stmtReady := `INSERT INTO dem_ready (order_num, creator, customer, dop_info, ms_note, creation_date) VALUES (?, ?, ?, ?, ?, ?)`

	creationDate := time.Date(fixture.Year, time.Month(fixture.Month), 15, 10, 0, 0, 0, time.UTC).Unix()
	var id int64

	result, err := testDB.Exec(stmtReady, fixture.OrderNum, fixture.Creator, fixture.Customer, fixture.DopInfo, fixture.MsNote, creationDate)
	require.NoError(t, err)

	id, err = result.LastInsertId()
	require.NoError(t, err)

	// 2. Вставляем позиции
	for _, pos := range fixture.Positions {
		_, err := testDB.Exec(`
			INSERT INTO dem_price (numorders, position, name_position, kol_vo)
			VALUES (?, ?, ?, ?)
		`, fixture.OrderNum, pos.Position, pos.NamePosition, pos.Count)
		require.NoError(t, err)
	}

	// 3. Вставляем изображения
	for _, img := range fixture.Images {
		_, err := testDB.Exec(`
			INSERT INTO dem_images (im_ordername, im_orderpos, im_image)
			VALUES (?, ?, ?)
		`, img.OrderName, img.OrderPos, img.Image)
		require.NoError(t, err)
	}

	// 4. Вставляем планы
	for _, plan := range fixture.Plans {
		_, err := testDB.Exec(`
			INSERT INTO dem_plan (idorder, x, color, sqr)
			VALUES (?, ?, ?, ?)
		`, plan.IDOrder, plan.X, plan.Color, plan.Sqr)
		require.NoError(t, err)
	}

	return id
}

func cleanupTestDB(t *testing.T) {
	tables := []string{"dem_ready", "dem_price", "dem_images", "dem_plan"}
	for _, table := range tables {
		_, err := testDB.Exec("DELETE FROM " + table)
		require.NoError(t, err)
	}
}

//func TestCreateOrder(t *testing.T) {
//	createTestOrderDem(t, "Q6-123", 2026, 01)
//}

//func TestDeleteOrder(t *testing.T) {
//	cleanupTestDB(t)
//}

func TestStorage_GetOrdersMonth(t *testing.T) {
	cleanupTestDB(t)

	// Ожидаемые номера заказов
	expectedOrderNums := []string{"Q6-0", "Q6-1", "Q6-2"}

	//pkg := TestOrderFixture{}

	for i := 0; i < 3; i++ {
		createTestOrderDem(t, TestOrderFixture{
			OrderNum:  "Q6-" + strconv.Itoa(i),
			Creator:   i,
			Customer:  "testCust" + strconv.Itoa(i),
			DopInfo:   "testDop" + strconv.Itoa(i),
			MsNote:    "testNot" + strconv.Itoa(i),
			Year:      2026,
			Month:     1,
			Positions: nil,
			Images:    nil,
			Plans:     nil,
		})
	}

	//for _, num := range expectedOrderNums {
	//	createTestOrderDem(t) // год — int, лучше 2026
	//}

	s := &Storage{db: testDB}
	orders, err := s.GetOrdersMonth(context.Background(), 2026, 1, "")
	require.NoError(t, err)
	assert.Len(t, orders, 3)

	// Собираем фактические номера
	actualOrderNums := make([]string, len(orders))
	for i, order := range orders {
		actualOrderNums[i] = order.OrderNum
	}

	// Проверяем, что все ожидаемые заказы получены (порядок не важен)
	assert.ElementsMatch(t, expectedOrderNums, actualOrderNums)
}

func TestStorage_GetOrdersMonthWithSearch(t *testing.T) {
	cleanupTestDB(t)

	for i := 0; i < 3; i++ {
		createTestOrderDem(t, TestOrderFixture{
			OrderNum:  "Q6-" + strconv.Itoa(i),
			Creator:   i,
			Customer:  "testCust" + strconv.Itoa(i),
			DopInfo:   "testDop" + strconv.Itoa(i),
			MsNote:    "testNot" + strconv.Itoa(i),
			Year:      2026,
			Month:     1,
			Positions: nil,
			Images:    nil,
			Plans:     nil,
		})
	}

	search := "Q6-0"

	s := &Storage{db: testDB}
	orders, err := s.GetOrdersMonth(context.Background(), 2026, 1, search)
	require.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, "Q6-0", orders[0].OrderNum)
}

func TestStorage_GetOrdersMonth_NoOrdersInMonth(t *testing.T) {
	cleanupTestDB(t)

	for i := 0; i < 3; i++ {
		createTestOrderDem(t, TestOrderFixture{
			OrderNum:  "Q6-" + strconv.Itoa(i),
			Creator:   i,
			Customer:  "testCust" + strconv.Itoa(i),
			DopInfo:   "testDop" + strconv.Itoa(i),
			MsNote:    "testNot" + strconv.Itoa(i),
			Year:      2026,
			Month:     2,
			Positions: nil,
			Images:    nil,
			Plans:     nil,
		})
	}

	s := &Storage{db: testDB}
	orders, err := s.GetOrdersMonth(context.Background(), 2026, 1, "") // январь
	require.NoError(t, err)
	assert.Empty(t, orders, "Ожидался пустой список заказов за январь 2026")
}

func TestStorage_GetOrdersMonth_SearchNotFound(t *testing.T) {
	cleanupTestDB(t)

	for i := 0; i < 3; i++ {
		createTestOrderDem(t, TestOrderFixture{
			OrderNum:  "Q6-" + strconv.Itoa(i),
			Creator:   i,
			Customer:  "testCust" + strconv.Itoa(i),
			DopInfo:   "testDop" + strconv.Itoa(i),
			MsNote:    "testNot" + strconv.Itoa(i),
			Year:      2026,
			Month:     1,
			Positions: nil,
			Images:    nil,
			Plans:     nil,
		})
	}

	s := &Storage{db: testDB}
	orders, err := s.GetOrdersMonth(context.Background(), 2026, 1, "NONEXISTENT")
	require.NoError(t, err)
	assert.Empty(t, orders, "Ожидался пустой результат при поиске несуществующего заказа")
}
