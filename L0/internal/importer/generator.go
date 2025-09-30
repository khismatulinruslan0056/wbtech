package importer

import (
	"bytes"
	"fmt"
	"math/rand"
	"text/template"
	"time"
)

var SupportedLocales = []string{
	"aa", "ab", "ae", "af", "ak", "am", "an", "ar", "as", "av", "ay", "az",
	"ba", "be", "bg", "bh", "bi", "bm", "bn", "bo", "br", "bs",
	"ca", "ce", "ch", "co", "cr", "cs", "cu", "cv", "cy",
	"da", "de", "dv", "dz",
	"ee", "el", "en", "eo", "es", "et", "eu",
	"fa", "ff", "fi", "fj", "fo", "fr", "fy",
	"ga", "gd", "gl", "gn", "gu", "gv",
	"ha", "he", "hi", "ho", "hr", "ht", "hu", "hy", "hz",
	"ia", "id", "ie", "ig", "ii", "ik", "io", "is", "it", "iu",
	"ja", "jv", "ka",
	"kg", "ki", "kj", "kk", "kl", "km", "kn", "ko", "kr", "ks", "ku", "kv", "kw",
	"ky", "la",
	"lb", "lg", "li", "ln", "lo", "lt", "lu", "lv",
	"mg", "mh", "mi", "mk", "ml", "mn", "mr", "ms", "mt", "my",
	"na", "nb", "nd", "ne", "ng", "nl", "nn", "no", "nr", "nv", "ny",
	"oc", "oj", "om", "or", "os",
	"pa", "pi", "pl", "ps", "pt",
	"qu", "rm", "rn", "ro", "ru", "rw",
	"sa", "sc", "sd", "se", "sg", "si", "sk", "sl", "sm", "sn", "so", "sq", "sr", "ss", "st", "su", "sv", "sw",
	"ta", "te", "tg", "th", "ti", "tk", "tl", "tn", "to", "tr", "ts", "tt", "tw", "ty",
	"ug", "uk", "ur", "uz",
	"ve", "vi", "vo",
	"wa", "wo",
	"xh", "yi", "yo",
	"za", "zh", "zu",
}

type Generator struct {
	rnd  *rand.Rand
	tmpl *template.Template
}

func NewGenerator(r *rand.Rand, tmpl *template.Template) *Generator {
	return &Generator{rnd: r, tmpl: tmpl}
}

var (
	letters         = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+")
	lettersForEmail = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func (g *Generator) randString(n int, valid bool) string {
	if !valid || n <= 0 {
		return ""
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[g.rnd.Intn(len(letters))]
	}
	return string(b)
}

func (g *Generator) randStringForEmail(n int, valid bool) string {
	if !valid || n <= 0 {
		return ""
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[g.rnd.Intn(len(lettersForEmail))]
	}
	return string(b)
}

func (g *Generator) randEmail(valid bool) string {
	if !valid {
		return g.randString(12, false)
	}
	return fmt.Sprintf("%s@test.ru", g.randStringForEmail(5, true))
}

func (g *Generator) randPhone(valid bool) string {
	if !valid {
		return g.randString(6, false)
	}
	return fmt.Sprintf("+%d%d", g.rnd.Intn(99), g.rnd.Intn(999999999))
}

func (g *Generator) randDate(valid bool) string {
	if !valid {
		badDates := []string{"not-a-date", "2021-99-99", "0000-00-00T00:00:00Z", "32-13-2025"}
		return badDates[g.rnd.Intn(len(badDates))]
	}
	return time.Now().Add(time.Duration(g.rnd.Intn(1000)) * time.Hour * -1).Format(time.RFC3339)
}

func (g *Generator) generate(i int, valid bool) (string, *Order) {
	trackNumber := fmt.Sprintf("TRK-%s", g.randString(5, true))
	phone := g.randPhone(true)
	date := g.randDate(true)
	email := g.randEmail(true)
	name := g.randString(10, true)
	city := g.randString(6, true)
	address := g.randString(15, true)
	region := g.randString(8, true)
	requestID := g.randString(5, true)
	bank := g.randString(6, true)
	itemTrackNumber := fmt.Sprintf("ITM-%s", g.randString(4, true))
	rid := g.randString(12, true)
	itemName := g.randString(7, true)
	brand := g.randString(8, true)
	locale := SupportedLocales[g.rnd.Intn(len(SupportedLocales))]
	internalSignature := g.randString(5, true)
	customerID := g.randString(6, true)
	shardKey := fmt.Sprintf("%d", g.rnd.Intn(10))
	oofShard := fmt.Sprintf("%d", g.rnd.Intn(5))
	zip := fmt.Sprintf("%d", g.rnd.Intn(999999))
	amount := g.rnd.Intn(20000)
	deliveryCost := g.rnd.Intn(5000)
	goodsTotal := g.rnd.Intn(5000)
	customFee := g.rnd.Intn(1000)
	chrtID := g.rnd.Intn(9999999)
	price := g.rnd.Intn(2000)
	sale := g.rnd.Intn(70)
	size := fmt.Sprintf("%d", g.rnd.Intn(5))
	totalPrice := g.rnd.Intn(5000)
	nmID := g.rnd.Intn(9999999)
	status := g.rnd.Intn(500)
	smID := g.rnd.Intn(1000)

	orderUID := fmt.Sprintf("%s-%d", g.randString(16, true), i)
	if !valid {
		switch i % 33 {
		case 0:
			phone = g.randPhone(valid)
		case 1:
			date = g.randDate(valid)
		case 2:
			email = g.randEmail(valid)
		case 3:
			name = g.randString(10, valid)
		case 4:
			city = g.randString(6, valid)
		case 5:
			address = g.randString(15, valid)
		case 6:
			region = g.randString(8, valid)
		case 7:
			requestID = g.randString(5, valid)
		case 8:
			bank = g.randString(6, valid)
		case 9:
			itemTrackNumber = fmt.Sprintf("ITM-%s", g.randString(4, valid))
		case 10:
			rid = g.randString(12, valid)
		case 11:
			itemName = g.randString(7, valid)
		case 12:
			brand = g.randString(8, valid)
		case 13:
			locale = "asd"
		case 14:
			internalSignature = g.randString(5, valid)
		case 15:
			customerID = g.randString(6, valid)
		case 16:
			shardKey = "-1"
		case 17:
			oofShard = "-1"
		case 18:
			trackNumber = ""
		case 19:
			zip = ""
		case 20:
			amount = -1
		case 21:
			deliveryCost = -1
		case 22:
			goodsTotal = -1
		case 23:
			customFee = -1
		case 24:
			chrtID = -1
		case 25:
			price = -1
		case 26:
			sale = -1
		case 27:
			size = ""
		case 28:
			totalPrice = -1
		case 29:
			nmID = -1
		case 30:
			status = -1
		case 31:
			smID = -1
		case 32:
			orderUID = ""
		}
	}

	return orderUID, &Order{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    name,
			Phone:   phone,
			Zip:     zip,
			City:    city,
			Address: address,
			Region:  region,
			Email:   email,
		},
		Payment: Payment{
			Transaction:  orderUID,
			RequestID:    requestID,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       amount,
			PaymentDT:    time.Now().Unix(),
			Bank:         bank,
			DeliveryCost: deliveryCost,
			GoodsTotal:   goodsTotal,
			CustomFee:    customFee,
		},
		Items: []Item{
			{
				ChrtID:      chrtID,
				TrackNumber: itemTrackNumber,
				Price:       price,
				Rid:         rid,
				Name:        itemName,
				Sale:        sale,
				Size:        size,
				TotalPrice:  totalPrice,
				NmID:        nmID,
				Brand:       brand,
				Status:      status,
			},
		},
		Locale:            locale,
		InternalSignature: internalSignature,
		CustomerID:        customerID,
		DeliveryService:   "meest",
		ShardKey:          shardKey,
		SmID:              smID,
		DateCreated:       date,
		OofShard:          oofShard,
	}
}

func (g *Generator) GenerateOrder(i int, valid bool) ([]byte, []byte, error) {
	const op = "GenerateOrder"
	k, order := g.generate(i, valid)
	var buf bytes.Buffer
	err := g.tmpl.Execute(&buf, order)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return []byte(k), buf.Bytes(), nil
}
