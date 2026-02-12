package services

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	docxgo "github.com/mmonterroca/docxgo"
	"github.com/mmonterroca/docxgo/domain"
)

const DateFormat = "2006-01-02"

type ReportingService struct {
	db *sqlx.DB
}

type ReportData struct {
	StartDate        time.Time
	EndDate          time.Time
	TotalRequests    int
	BlockedRequests  int
	SuccessRate      float64
	TopIPs           []StatItem
	TopCountries     []StatItem
	RecentAttacks    []AttackLog
	TopAttackTypes   []AttackTypeStat
	DailyTraffic     []DailyStat
	TopAttackVectors []StatItem
}

type DailyStat struct {
	Date    string
	Total   int
	Blocked int
}

type StatItem struct {
	Name        string
	CountryCode string
	CountryName string
	Count       int
}

type AttackTypeStat struct {
	Country    string
	AttackType string
	Count      int
}

type AttackLog struct {
	Time    string
	IP      string
	Type    string
	Country string
}

type Translations struct {
	Title                   string
	Summary                 string
	TotalRequests           string
	BlockedAttacks          string
	TrafficHealth           string
	TopIPs                  string
	TopCountries            string
	RecentAttacks           string
	Conclusion              string
	Recommendations         string
	Recommendation1         string
	Recommendation2         string
	Recommendation3         string
	IPAddress               string
	Violations              string
	Time                    string
	Country                 string
	Type                    string
	NoAttacks               string
	NoRecent                string
	ConclusionText          string
	NoThreatsText           string
	TopAttackTypes          string
	TrafficHealthExpHealthy string
	TrafficHealthExpWarn    string
	AttackChartExp          string
	TrafficVolumeTitle      string // New
	TrafficVolumeDesc       string // New
	TopAttackVectorsTitle   string // New
	TopAttackVectorsDesc    string // New
	TrafficClassTitle       string // New
	TopSourcesTitle         string // New
	Last24Hours             string // New
	Last7Days               string // New
	Last30Days              string // New
	CustomRange             string // New
}

var translations = map[string]Translations{
	"en": {
		Title:                   "DoCode WAF Security Report",
		Summary:                 "Summary Statistics",
		TotalRequests:           "Total Requests",
		BlockedAttacks:          "Blocked Attacks",
		TrafficHealth:           "Traffic Health",
		TopIPs:                  "Top Attacking IPs",
		TopCountries:            "Top Countries",
		RecentAttacks:           "Recent Blocked Requests",
		Conclusion:              "Conclusion",
		Recommendations:         "Recommendations",
		Recommendation1:         "1. Review the Top Attacking IPs list and add permanent bans for persistent attackers.",
		Recommendation2:         "2. Check the 'Recent Blocked Requests' to identify any patterns targeting specific vulnerabilities.",
		Recommendation3:         "3. Ensure that all security rules are up to date and blocking mode is enabled for critical paths.",
		IPAddress:               "IP Address",
		Violations:              "Violations",
		Time:                    "Time",
		Country:                 "Country",
		Type:                    "Type",
		NoAttacks:               "No attacks recorded in this period.",
		NoRecent:                "No recent attacks found.",
		ConclusionText:          "During this period, the WAF processed %d requests and blocked %d potential threats. The system maintained a traffic success rate of %.2f%%. ",
		NoThreatsText:           "No significant security threats were detected.",
		TopAttackTypes:          "Top Attack Types by Country",
		TrafficHealthExpHealthy: "The chart above illustrates the traffic composition. %.2f%% of requests were allowed, demonstrating healthy traffic flow.",
		TrafficHealthExpWarn:    "The chart indicates a high volume of blocked requests (%.2f%% success rate), suggesting active attack attempts.",
		AttackChartExp:          "The bar chart highlights that %s is the most frequent source of blocked requests, with %d violations recorded.",
		TrafficVolumeTitle:      "Traffic Volume & Threats (%s)",
		TrafficVolumeDesc:       "Daily trend of total vs blocked requests.",
		TopAttackVectorsTitle:   "Top 5 Attack Vectors Blocked",
		TopAttackVectorsDesc:    "Most common attack types blocked by the WAF.",
		TrafficClassTitle:       "Traffic Classification Distribution",
		TopSourcesTitle:         "Top 5 Sources of Attack by Country",
		Last24Hours:             "Last 24 Hours",
		Last7Days:               "Last 7 Days",
		Last30Days:              "Last 30 Days",
		CustomRange:             "Custom Range",
	},
	"id": {
		Title:                   "Laporan Keamanan WAF DoCode",
		Summary:                 "Statistik Ringkasan",
		TotalRequests:           "Total Permintaan",
		BlockedAttacks:          "Serangan Diblokir",
		TrafficHealth:           "Kesehatan Lalu Lintas",
		TopIPs:                  "IP Penyerang Teratas",
		TopCountries:            "Negara Teratas",
		RecentAttacks:           "Permintaan Diblokir Terbaru",
		Conclusion:              "Kesimpulan",
		Recommendations:         "Rekomendasi",
		Recommendation1:         "1. Tinjau daftar IP Penyerang Teratas dan tambahkan larangan permanen untuk penyerang persisten.",
		Recommendation2:         "2. Periksa 'Permintaan Diblokir Terbaru' untuk mengidentifikasi pola yang menargetkan kerentanan tertentu.",
		Recommendation3:         "3. Pastikan semua aturan keamanan sudah diperbarui dan mode pemblokiran diaktifkan untuk jalur kritis.",
		IPAddress:               "Alamat IP",
		Violations:              "Pelanggaran",
		Time:                    "Waktu",
		Country:                 "Negara",
		Type:                    "Tipe",
		NoAttacks:               "Tidak ada serangan tercatat pada periode ini.",
		NoRecent:                "Tidak ada serangan terbaru ditemukan.",
		ConclusionText:          "Selama periode ini, WAF memproses %d permintaan dan memblokir %d ancaman potensial. Sistem mempertahankan tingkat keberhasilan lalu lintas sebesar %.2f%%. ",
		NoThreatsText:           "Tidak ada ancaman keamanan signifikan yang terdeteksi.",
		TopAttackTypes:          "Tipe Serangan Teratas berdasarkan Negara",
		TrafficHealthExpHealthy: "Grafik di atas menggambarkan komposisi lalu lintas. %.2f%% permintaan diizinkan, menunjukkan aliran lalu lintas yang sehat.",
		TrafficHealthExpWarn:    "Grafik menunjukkan volume tinggi permintaan yang diblokir (tingkat keberhasilan %.2f%%), mengindikasikan upaya serangan aktif.",
		AttackChartExp:          "Grafik batang menyoroti bahwa %s adalah sumber permintaan yang diblokir paling sering, dengan %d pelanggaran tercatat.",
		TrafficVolumeTitle:      "Volume Lalu Lintas & Ancaman (%s)",
		TrafficVolumeDesc:       "Tren harian total permintaan vs yang diblokir.",
		TopAttackVectorsTitle:   "5 Vektor Serangan Teratas yang Diblokir",
		TopAttackVectorsDesc:    "Jenis serangan paling umum yang diblokir oleh WAF.",
		TrafficClassTitle:       "Distribusi Klasifikasi Lalu Lintas",
		TopSourcesTitle:         "5 Sumber Serangan Teratas berdasarkan Negara",
		Last24Hours:             "24 Jam Terakhir",
		Last7Days:               "7 Hari Terakhir",
		Last30Days:              "30 Hari Terakhir",
		CustomRange:             "Rentang Kustom",
	},
}

func getCountryName(code string) string {
	countries := map[string]string{
		"US": "United States", "CN": "China", "RU": "russia", "IN": "India",
		"BR": "Brazil", "DE": "Germany", "GB": "United Kingdom", "FR": "France",
		"ID": "Indonesia", "SG": "Singapore", "JP": "Japan", "KR": "South Korea",
		"VN": "Vietnam", "TH": "Thailand", "MY": "Malaysia", "PH": "Philippines",
		"AU": "Australia", "CA": "Canada", "NL": "Netherlands", "TR": "Turkey",
		"UA": "Ukraine", "IR": "Iran", "Unknown": "Unknown",
	}
	if name, ok := countries[code]; ok {
		return name
	}
	return code
}

func NewReportingService(db *sqlx.DB) *ReportingService {
	return &ReportingService{db: db}
}

func (s *ReportingService) GenerateReportData(start, end time.Time) (*ReportData, error) {
	data := &ReportData{
		StartDate: start,
		EndDate:   end,
	}

	// 1. Total Requests
	queryTotal := `SELECT COUNT(*) FROM traffic_logs WHERE timestamp >= $1 AND timestamp <= $2`
	if err := s.db.Get(&data.TotalRequests, queryTotal, start, end); err != nil {
		return nil, err
	}

	// 2. Blocked Requests
	queryBlocked := `SELECT COUNT(*) FROM traffic_logs WHERE timestamp >= $1 AND timestamp <= $2 AND blocked = true`
	if err := s.db.Get(&data.BlockedRequests, queryBlocked, start, end); err != nil {
		return nil, err
	}

	// Calculate Success Rate
	if data.TotalRequests > 0 {
		allowed := data.TotalRequests - data.BlockedRequests
		data.SuccessRate = (float64(allowed) / float64(data.TotalRequests)) * 100
	} else {
		data.SuccessRate = 100
	}

	// 3. Top IPs (Attackers)
	queryTopIPs := `
		SELECT client_ip as name, country_code as countrycode, COUNT(*) as count 
		FROM traffic_logs 
		WHERE timestamp >= $1 AND timestamp <= $2 AND blocked = true
		GROUP BY client_ip, country_code 
		ORDER BY count DESC 
		LIMIT 10
	`
	if err := s.db.Select(&data.TopIPs, queryTopIPs, start, end); err != nil {
		return nil, err
	}

	// Populate Country Names
	for i := range data.TopIPs {
		if data.TopIPs[i].CountryCode == "" {
			data.TopIPs[i].CountryCode = "Unknown"
		}
		data.TopIPs[i].CountryName = getCountryName(data.TopIPs[i].CountryCode)
	}

	// 4. Top Countries
	queryTopCountries := `
		SELECT COALESCE(NULLIF(country_code, ''), 'Unknown') as name, COUNT(*) as count 
		FROM traffic_logs 
		WHERE timestamp >= $1 AND timestamp <= $2 AND blocked = true
		GROUP BY name 
		ORDER BY count DESC 
		LIMIT 10
	`
	if err := s.db.Select(&data.TopCountries, queryTopCountries, start, end); err != nil {
		return nil, err
	}

	// Same for TopCountries
	for i := range data.TopCountries {
		if data.TopCountries[i].Name == "" {
			data.TopCountries[i].Name = "Unknown"
		}
		// Code is in Name field for TopCountries based on query
		data.TopCountries[i].CountryName = getCountryName(data.TopCountries[i].Name)
	}

	// 5. Recent Attacks
	queryRecent := `
		SELECT to_char(timestamp, 'YYYY-MM-DD HH24:MI:SS') as time, client_ip as ip, COALESCE(attack_type, 'Manual Block') as type, COALESCE(country_code, 'Unknown') as country
		FROM traffic_logs
		WHERE timestamp >= $1 AND timestamp <= $2 AND blocked = true
		ORDER BY timestamp DESC
		LIMIT 20
	`
	if err := s.db.Select(&data.RecentAttacks, queryRecent, start, end); err != nil {
		return nil, err
	}

	// 6. Top Attack Types by Country
	queryAttackTypes := `
		SELECT COALESCE(country_code, 'Unknown') as country, COALESCE(attack_type, 'Manual') as attacktype, COUNT(*) as count
		FROM traffic_logs
		WHERE timestamp >= $1 AND timestamp <= $2 AND blocked = true
		GROUP BY country_code, attack_type
		ORDER BY count DESC
		LIMIT 10
	`
	if err := s.db.Select(&data.TopAttackTypes, queryAttackTypes, start, end); err != nil {
		return nil, err
	}

	// Format country names in Attack Types
	for i := range data.TopAttackTypes {
		data.TopAttackTypes[i].Country = getCountryName(data.TopAttackTypes[i].Country)
	}

	// 7. Daily Traffic Stats (for Line Chart)
	queryDaily := `
		SELECT to_char(timestamp, 'YYYY-MM-DD') as date, 
		       COUNT(*) as total, 
		       COUNT(*) FILTER (WHERE blocked = true) as blocked
		FROM traffic_logs
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY date
		ORDER BY date ASC
	`
	var rawDaily []DailyStat
	if err := s.db.Select(&rawDaily, queryDaily, start, end); err != nil {
		return nil, err
	}

	// Fill missing dates
	dailyMap := make(map[string]DailyStat)
	for _, d := range rawDaily {
		dailyMap[d.Date] = d
	}

	data.DailyTraffic = []DailyStat{}
	current := start
	// Iterate through each day from start to end
	for !current.After(end) {
		dateStr := current.Format("2006-01-02")
		if stat, exists := dailyMap[dateStr]; exists {
			data.DailyTraffic = append(data.DailyTraffic, stat)
		} else {
			data.DailyTraffic = append(data.DailyTraffic, DailyStat{
				Date:    dateStr,
				Total:   0,
				Blocked: 0,
			})
		}
		current = current.AddDate(0, 0, 1)
	}

	// 8. Top Attack Vectors (Aggregated for Horizontal Bar Chart)
	queryVectors := `
		SELECT COALESCE(NULLIF(attack_type, ''), 'Manual Block') as name, COUNT(*) as count
		FROM traffic_logs
		WHERE timestamp >= $1 AND timestamp <= $2 AND blocked = true
		GROUP BY name
		ORDER BY count DESC
		LIMIT 5
	`
	if err := s.db.Select(&data.TopAttackVectors, queryVectors, start, end); err != nil {
		return nil, err
	}

	return data, nil
}

func getTimeRangeLabel(start, end time.Time, t Translations) string {
	duration := end.Sub(start)
	hours := duration.Hours()

	if hours <= 25 {
		return t.Last24Hours
	} else if hours > 24*6 && hours < 24*8 {
		return t.Last7Days
	} else if hours > 24*29 && hours < 24*31 {
		return t.Last30Days
	}
	return t.CustomRange
}

func (s *ReportingService) GeneratePDF(data *ReportData, lang string) ([]byte, error) {
	t := translations["en"]
	if val, ok := translations[lang]; ok {
		t = val
	}

	m := pdf.NewMaroto(consts.Landscape, consts.A4)
	m.SetPageMargins(20, 10, 20)

	// Header
	m.RegisterHeader(func() {
		m.Row(20, func() {
			m.Col(12, func() {
				m.Text(t.Title, props.Text{
					Size:  18,
					Style: consts.Bold,
					Align: consts.Center,
				})
				m.Text(fmt.Sprintf("%s - %s", data.StartDate.Format(DateFormat), data.EndDate.Format(DateFormat)), props.Text{
					Size:  10,
					Align: consts.Center,
					Top:   12,
				})
			})
		})
	})

	// Summary Stats
	m.Row(30, func() {
		m.Col(4, func() {
			m.Text(t.TotalRequests, props.Text{Size: 10, Align: consts.Center, Style: consts.Bold})
			m.Text(fmt.Sprintf("%d", data.TotalRequests), props.Text{Size: 14, Align: consts.Center, Top: 10})
		})
		m.Col(4, func() {
			m.Text(t.BlockedAttacks, props.Text{Size: 10, Align: consts.Center, Style: consts.Bold, Color: color.Color{Red: 255, Green: 0, Blue: 0}})
			m.Text(fmt.Sprintf("%d", data.BlockedRequests), props.Text{Size: 14, Align: consts.Center, Top: 10, Color: color.Color{Red: 255, Green: 0, Blue: 0}})
		})
		m.Col(4, func() {
			m.Text(t.TrafficHealth, props.Text{Size: 10, Align: consts.Center, Style: consts.Bold})
			m.Text(fmt.Sprintf("%.2f%%", data.SuccessRate), props.Text{Size: 14, Align: consts.Center, Top: 10})
		})
	})

	m.Line(10)

	// 1. Traffic Volume & Threats (Line Chart)
	if len(data.DailyTraffic) > 0 {
		lineChartBytes, err := generateTrafficLineChart(data.DailyTraffic)
		if err == nil {
			tmpFile, err := os.CreateTemp("", "pdf-chart-line-*.png")
			if err == nil {
				defer os.Remove(tmpFile.Name())
				if _, err := tmpFile.Write(lineChartBytes); err == nil {
					tmpFile.Close()

					timeRange := getTimeRangeLabel(data.StartDate, data.EndDate, t)

					m.Row(10, func() {
						m.Col(12, func() {
							m.Text(fmt.Sprintf(t.TrafficVolumeTitle, timeRange), props.Text{Size: 12, Style: consts.Bold, Align: consts.Center})
						})
					})
					m.Row(80, func() {
						m.Col(12, func() {
							m.FileImage(tmpFile.Name(), props.Rect{
								Center:  true,
								Percent: 80,
							})
						})
					})
					m.Row(10, func() {
						m.Col(12, func() {
							m.Text(t.TrafficVolumeDesc, props.Text{Size: 9, Style: consts.Italic, Align: consts.Center})
						})
					})
				}
			}
		}
	}

	m.Line(5)

	// 2. Top 5 Attack Vectors Blocked (Horizontal Bar Chart simulated)
	log.Printf("[Report] TopAttackVectors count: %d", len(data.TopAttackVectors))
	if len(data.TopAttackVectors) > 0 {
		hBarChartBytes, err := generateAttackVectorsBarChart(data.TopAttackVectors)
		log.Printf("[Report] Attack Vectors chart generation: bytes=%d, err=%v", len(hBarChartBytes), err)
		if err == nil {
			tmpFile, err := os.CreateTemp("", "pdf-chart-hbar-*.png")
			if err == nil {
				defer os.Remove(tmpFile.Name())
				if _, err := tmpFile.Write(hBarChartBytes); err == nil {
					tmpFile.Close()
					log.Printf("[Report] Attack Vectors chart written to: %s", tmpFile.Name())

					m.Row(10, func() {
						m.Col(12, func() {
							m.Text(t.TopAttackVectorsTitle, props.Text{Size: 12, Style: consts.Bold, Align: consts.Center})
						})
					})
					m.Row(80, func() {
						m.Col(12, func() {
							m.FileImage(tmpFile.Name(), props.Rect{
								Center:  true,
								Percent: 80,
							})
						})
					})
					m.Row(10, func() {
						m.Col(12, func() {
							m.Text(t.TopAttackVectorsDesc, props.Text{Size: 9, Style: consts.Italic, Align: consts.Center})
						})
					})
				} else {
					log.Printf("[Report] Attack Vectors chart write error: %v", err)
				}
			} else {
				log.Printf("[Report] Attack Vectors temp file error: %v", err)
			}
		}
	}

	m.Line(5)

	// 3. Top 5 Sources of Attack by Country
	log.Printf("[Report] TopCountries count: %d", len(data.TopCountries))
	barChartBytes, err := generateAttacksBarChart(data.TopCountries)
	log.Printf("[Report] TopCountries chart generation: bytes=%d, err=%v", len(barChartBytes), err)
	if err == nil {
		tmpFile, err := os.CreateTemp("", "pdf-chart-attacks-*.png")
		if err == nil {
			defer os.Remove(tmpFile.Name())
			if _, err := tmpFile.Write(barChartBytes); err == nil {
				tmpFile.Close()
				log.Printf("[Report] TopCountries chart written to: %s", tmpFile.Name())
				m.Row(10, func() {
					m.Col(12, func() {
						m.Text(t.TopSourcesTitle, props.Text{Size: 12, Style: consts.Bold, Align: consts.Center})
					})
				})
				m.Row(80, func() {
					m.Col(12, func() {
						m.FileImage(tmpFile.Name(), props.Rect{
							Center:  true,
							Percent: 80,
						})
					})
				})

				// Explanation for Bar Chart
				if len(data.TopCountries) > 0 {
					topCountry := data.TopCountries[0]
					barExp := fmt.Sprintf(t.AttackChartExp, topCountry.CountryName, topCountry.Count)
					m.Row(15, func() {
						m.Col(12, func() {
							m.Text(barExp, props.Text{
								Size:  9,
								Align: consts.Center,
								Style: consts.Italic,
								Top:   5,
							})
						})
					})
				}
			} else {
				log.Printf("[Report] TopCountries chart write error: %v", err)
			}
		} else {
			log.Printf("[Report] TopCountries temp file error: %v", err)
		}
	}

	m.Line(5)

	// 4. Traffic Classification Distribution (Donut Chart)
	allowed := data.TotalRequests - data.BlockedRequests
	donutChartBytes, err := generateTrafficDonutChart(allowed, data.BlockedRequests)
	if err == nil {
		tmpFile, err := os.CreateTemp("", "pdf-chart-donut-*.png")
		if err == nil {
			defer os.Remove(tmpFile.Name())
			if _, err := tmpFile.Write(donutChartBytes); err == nil {
				tmpFile.Close()
				m.Row(10, func() {
					m.Col(12, func() {
						m.Text(t.TrafficClassTitle, props.Text{Size: 12, Style: consts.Bold, Align: consts.Center})
					})
				})
				m.Row(80, func() {
					m.Col(12, func() {
						m.FileImage(tmpFile.Name(), props.Rect{
							Center:  true,
							Percent: 80,
						})
					})
				})
				// Explanation
				explanation := fmt.Sprintf(t.TrafficHealthExpHealthy, data.SuccessRate)
				if data.SuccessRate < 90 {
					explanation = fmt.Sprintf(t.TrafficHealthExpWarn, data.SuccessRate)
				}
				m.Row(15, func() {
					m.Col(12, func() {
						m.Text(explanation, props.Text{
							Size:  9,
							Align: consts.Center,
							Style: consts.Italic,
							Top:   5,
						})
					})
				})
			}
		}
	}

	m.Line(10)

	// Top IPs Table
	m.Row(10, func() {
		m.Col(12, func() {
			m.Text(t.TopIPs, props.Text{Size: 12, Style: consts.Bold, Top: 2})
		})
	})

	headers := []string{t.IPAddress, t.Country, t.Violations}
	var contents [][]string
	for _, ip := range data.TopIPs {
		contents = append(contents, []string{ip.Name, fmt.Sprintf("%s (%s)", ip.CountryName, ip.CountryCode), fmt.Sprintf("%d", ip.Count)})
	}

	if len(contents) > 0 {
		m.TableList(headers, contents, props.TableList{
			HeaderProp: props.TableListContent{
				Size:      10,
				GridSizes: []uint{5, 5, 2},
			},
			ContentProp: props.TableListContent{
				Size:      10,
				GridSizes: []uint{5, 5, 2},
			},
			Align:                consts.Left,
			AlternatedBackground: &color.Color{Red: 240, Green: 240, Blue: 240},
			Line:                 true,
		})
	} else {
		m.Row(10, func() {
			m.Col(12, func() {
				m.Text(t.NoAttacks, props.Text{Style: consts.Italic})
			})
		})
	}

	m.Row(10, func() {}) // Spacer

	// Top Attack Types by Country
	m.Row(10, func() {
		m.Col(12, func() {
			m.Text(t.TopAttackTypes, props.Text{Size: 12, Style: consts.Bold, Top: 2})
		})
	})

	// Generate Bar Chart for Attack Types

	typeHeaders := []string{t.Country, t.Type, t.Violations}
	var typeContents [][]string
	for _, stat := range data.TopAttackTypes {
		typeContents = append(typeContents, []string{stat.Country, stat.AttackType, fmt.Sprintf("%d", stat.Count)})
	}

	if len(typeContents) > 0 {
		m.TableList(typeHeaders, typeContents, props.TableList{
			HeaderProp: props.TableListContent{
				Size:      10,
				GridSizes: []uint{5, 5, 2},
			},
			ContentProp: props.TableListContent{
				Size:      10,
				GridSizes: []uint{5, 5, 2},
			},
			Align:                consts.Left,
			AlternatedBackground: &color.Color{Red: 240, Green: 240, Blue: 240},
			Line:                 true,
		})
	}

	m.Row(10, func() {}) // Spacer

	// Recent Attacks Table
	m.Row(10, func() {
		m.Col(12, func() {
			m.Text(t.RecentAttacks, props.Text{Size: 12, Style: consts.Bold, Top: 2})
		})
	})

	recentHeaders := []string{t.Time, t.IPAddress, t.Country, t.Type}
	var recentContents [][]string
	for _, attack := range data.RecentAttacks {
		recentContents = append(recentContents, []string{attack.Time, attack.IP, getCountryName(attack.Country), attack.Type})
	}

	if len(recentContents) > 0 {
		m.TableList(recentHeaders, recentContents, props.TableList{
			HeaderProp: props.TableListContent{
				Size:      9,
				GridSizes: []uint{3, 3, 2, 4},
			},
			ContentProp: props.TableListContent{
				Size:      9,
				GridSizes: []uint{3, 3, 2, 4},
			},
			Align:                consts.Left,
			AlternatedBackground: &color.Color{Red: 240, Green: 240, Blue: 240},
			Line:                 true,
		})
	} else {
		m.Row(10, func() {
			m.Col(12, func() {
				m.Text(t.NoRecent, props.Text{Style: consts.Italic})
			})
		})
	}

	m.Row(20, func() {}) // Spacer

	// Conclusion & Recommendations
	m.Line(2)

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text(t.Conclusion, props.Text{Size: 12, Style: consts.Bold})
		})
	})

	conclusionText := fmt.Sprintf(t.ConclusionText, data.TotalRequests, data.BlockedRequests, data.SuccessRate)

	if data.BlockedRequests > 0 {
		if lang == "id" {
			conclusionText += " Aktivitas berbahaya terdeteksi dan dimitigasi sesuai dengan aturan keamanan yang aktif."
		} else {
			conclusionText += " Malicious activity was detected and mitigated according to the active security rules."
		}
	} else {
		conclusionText += t.NoThreatsText
	}

	m.Row(20, func() {
		m.Col(12, func() {
			m.Text(conclusionText, props.Text{Size: 10})
		})
	})

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text(t.Recommendations, props.Text{Size: 12, Style: consts.Bold, Top: 5})
		})
	})

	m.Row(30, func() {
		m.Col(12, func() {
			m.Text(t.Recommendation1, props.Text{Size: 10})
			m.Text(t.Recommendation2, props.Text{Size: 10, Top: 5})
			m.Text(t.Recommendation3, props.Text{Size: 10, Top: 10})
		})
	})

	buff, err := m.Output()
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil

}

func (s *ReportingService) GenerateDOCX(data *ReportData, lang string) ([]byte, error) {
	t := translations["en"]
	if val, ok := translations[lang]; ok {
		t = val
	}

	// Use DocumentBuilder for fluent API with landscape orientation
	builder := docxgo.NewDocumentBuilder()
	// Set the default section to landscape orientation
	builder.DefaultSection().Orientation(domain.OrientationLandscape).End()

	// Title
	builder.AddParagraph().Text(t.Title).FontSize(36).Bold().Alignment(docxgo.AlignmentCenter).End()
	builder.AddParagraph().Text(fmt.Sprintf("%s - %s", data.StartDate.Format(DateFormat), data.EndDate.Format(DateFormat))).FontSize(20).Alignment(docxgo.AlignmentCenter).End()
	builder.AddParagraph().End() // Spacer

	// Summary Section
	builder.AddParagraph().Text(t.Summary).FontSize(24).Bold().End()

	// 1. Traffic Volume & Threats (Line Chart)
	if len(data.DailyTraffic) > 0 {
		lineChartBytes, err := generateTrafficLineChart(data.DailyTraffic)
		if err == nil {
			tmpFile, err := os.CreateTemp("", "chart-line-*.png")
			if err == nil {
				defer os.Remove(tmpFile.Name())
				if _, err := tmpFile.Write(lineChartBytes); err == nil {
					tmpFile.Close()
					timeRange := getTimeRangeLabel(data.StartDate, data.EndDate, t)
					builder.AddParagraph().Text(fmt.Sprintf(t.TrafficVolumeTitle, timeRange)).FontSize(12).Bold().Alignment(docxgo.AlignmentCenter).End()
					builder.AddParagraph().AddImage(tmpFile.Name()).Alignment(docxgo.AlignmentCenter).End()
					builder.AddParagraph().Text(t.TrafficVolumeDesc).FontSize(10).Italic().Alignment(docxgo.AlignmentCenter).End()
					builder.AddParagraph().End()
				}
			}
		}
	}

	// 2. Top 5 Attack Vectors Blocked (Horizontal Bar Chart simulated)
	if len(data.TopAttackVectors) > 0 {
		hBarChartBytes, err := generateAttackVectorsBarChart(data.TopAttackVectors)
		if err == nil {
			tmpFile, err := os.CreateTemp("", "chart-hbar-*.png")
			if err == nil {
				defer os.Remove(tmpFile.Name())
				if _, err := tmpFile.Write(hBarChartBytes); err == nil {
					tmpFile.Close()
					builder.AddParagraph().Text(t.TopAttackVectorsTitle).FontSize(12).Bold().Alignment(docxgo.AlignmentCenter).End()
					builder.AddParagraph().AddImage(tmpFile.Name()).Alignment(docxgo.AlignmentCenter).End()
					builder.AddParagraph().Text(t.TopAttackVectorsDesc).FontSize(10).Italic().Alignment(docxgo.AlignmentCenter).End()
					builder.AddParagraph().End()
				}
			}
		}
	}

	// 3. Traffic Classification Distribution (Donut Chart) & 4. Top 5 Sources by Country
	// We can place them side by side if possible, or one after another. DOCX table can helper side-by-side.
	// For simplicity, let's stack them.

	// Top 5 Sources of Attack by Country (Existing Bar Chart logic but using TopCountries)
	// User reference shows "Top 5 Sources of Attack by Country". We have TopCountries data.
	barChartBytes, err := generateAttacksBarChart(data.TopCountries)
	if err == nil {
		tmpFile, err := os.CreateTemp("", "chart-attacks-*.png")
		if err == nil {
			defer os.Remove(tmpFile.Name())
			if _, err := tmpFile.Write(barChartBytes); err == nil {
				tmpFile.Close()
				builder.AddParagraph().Text(t.TopSourcesTitle).FontSize(12).Bold().Alignment(docxgo.AlignmentCenter).End()
				builder.AddParagraph().AddImage(tmpFile.Name()).Alignment(docxgo.AlignmentCenter).End()
				if len(data.TopCountries) > 0 {
					topCountry := data.TopCountries[0]
					explanation := fmt.Sprintf(t.AttackChartExp, topCountry.CountryName, topCountry.Count)
					builder.AddParagraph().Text(explanation).FontSize(10).Italic().Alignment(docxgo.AlignmentCenter).End()
				}
				builder.AddParagraph().End()
			}
		}
	}

	// Traffic Classification Distribution (Donut Chart)
	// Replacing previous Pie Chart section logic
	allowed := data.TotalRequests - data.BlockedRequests
	donutChartBytes, err := generateTrafficDonutChart(allowed, data.BlockedRequests)
	if err == nil {
		tmpFile, err := os.CreateTemp("", "chart-donut-*.png")
		if err == nil {
			defer os.Remove(tmpFile.Name())
			if _, err := tmpFile.Write(donutChartBytes); err == nil {
				tmpFile.Close()
				builder.AddParagraph().Text(t.TrafficClassTitle).FontSize(12).Bold().Alignment(docxgo.AlignmentCenter).End()
				builder.AddParagraph().AddImage(tmpFile.Name()).Alignment(docxgo.AlignmentCenter).End()

				// Explanation
				explanation := fmt.Sprintf(t.TrafficHealthExpHealthy, data.SuccessRate)
				if data.SuccessRate < 90 {
					explanation = fmt.Sprintf(t.TrafficHealthExpWarn, data.SuccessRate)
				}
				builder.AddParagraph().Text(explanation).FontSize(10).Italic().Alignment(docxgo.AlignmentCenter).End()
				builder.AddParagraph().End()
			}
		}
	}

	// Stats Text (Summary Metrics)
	builder.AddParagraph().Text(fmt.Sprintf("%s: %d", t.TotalRequests, data.TotalRequests)).End()
	builder.AddParagraph().Text(fmt.Sprintf("%s: %d", t.BlockedAttacks, data.BlockedRequests)).Color(docxgo.Red).End()
	builder.AddParagraph().Text(fmt.Sprintf("%s: %.2f%%", t.TrafficHealth, data.SuccessRate)).End()
	builder.AddParagraph().End()

	// Top IPs Table (Existing)
	builder.AddParagraph().Text(t.TopIPs).FontSize(24).Bold().End()
	if len(data.TopIPs) > 0 {
		tb := builder.AddTable(len(data.TopIPs)+1, 3).Style(domain.TableStyleGrid)
		// Header
		tb.Row(0).Cell(0).Text(t.IPAddress).Bold().End().
			Cell(1).Text(t.Country).Bold().End().
			Cell(2).Text(t.Violations).Bold().End().End()

		for i, ip := range data.TopIPs {
			tb.Row(i + 1).Cell(0).Text(ip.Name).End().
				Cell(1).Text(fmt.Sprintf("%s (%s)", ip.CountryName, ip.CountryCode)).End().
				Cell(2).Text(fmt.Sprintf("%d", ip.Count)).End().End()
		}
		tb.End()
	} else {
		builder.AddParagraph().Text(t.NoAttacks).Italic().End()
	}
	builder.AddParagraph().End()

	// Top Attack Types by Country Table (Existing)
	builder.AddParagraph().Text(t.TopAttackTypes).FontSize(24).Bold().End()
	if len(data.TopAttackTypes) > 0 {
		tb := builder.AddTable(len(data.TopAttackTypes)+1, 3).Style(domain.TableStyleGrid)
		tb.Row(0).Cell(0).Text(t.Country).Bold().End().
			Cell(1).Text(t.Type).Bold().End().
			Cell(2).Text(t.Violations).Bold().End().End()

		for i, stat := range data.TopAttackTypes {
			tb.Row(i + 1).Cell(0).Text(stat.Country).End().
				Cell(1).Text(stat.AttackType).End().
				Cell(2).Text(fmt.Sprintf("%d", stat.Count)).End().End()
		}
		tb.End()
	} else {
		builder.AddParagraph().Text(t.NoAttacks).Italic().End()
	}
	builder.AddParagraph().End()

	// Recent Attacks
	builder.AddParagraph().Text(t.RecentAttacks).FontSize(24).Bold().End()
	if len(data.RecentAttacks) > 0 {
		tb := builder.AddTable(len(data.RecentAttacks)+1, 4).Style(domain.TableStyleGrid)
		tb.Row(0).Cell(0).Text(t.Time).Bold().End().
			Cell(1).Text(t.IPAddress).Bold().End().
			Cell(2).Text(t.Country).Bold().End().
			Cell(3).Text(t.Type).Bold().End().End()

		for i, attack := range data.RecentAttacks {
			tb.Row(i + 1).Cell(0).Text(attack.Time).End().
				Cell(1).Text(attack.IP).End().
				Cell(2).Text(getCountryName(attack.Country)).End().
				Cell(3).Text(attack.Type).End().End()
		}
		tb.End()
	} else {
		builder.AddParagraph().Text(t.NoRecent).Italic().End()
	}
	builder.AddParagraph().End()

	// Conclusion
	builder.AddParagraph().Text(t.Conclusion).FontSize(24).Bold().End()
	conclusionText := fmt.Sprintf(t.ConclusionText, data.TotalRequests, data.BlockedRequests, data.SuccessRate)
	if data.BlockedRequests > 0 {
		if lang == "id" {
			conclusionText += " Aktivitas berbahaya terdeteksi dan dimitigasi sesuai dengan aturan keamanan yang aktif."
		} else {
			conclusionText += " Malicious activity was detected and mitigated according to the active security rules."
		}
	} else {
		conclusionText += t.NoThreatsText
	}
	builder.AddParagraph().Text(conclusionText).End()
	builder.AddParagraph().End()

	// Recommendations
	builder.AddParagraph().Text(t.Recommendations).FontSize(24).Bold().End()
	builder.AddParagraph().Text(t.Recommendation1).End()
	builder.AddParagraph().Text(t.Recommendation2).End()
	builder.AddParagraph().Text(t.Recommendation3).End()

	// Build Document
	doc, err := builder.Build()
	if err != nil {
		return nil, err
	}

	// Post-processing: Apply borders to all tables manually
	// because TableStyleGrid might not be rendering as expected in all viewers
	border := domain.BorderStyle{
		Style: domain.BorderSingle,
		Width: 4,                              // 0.5pt (4 eighths of a point)
		Color: domain.Color{R: 0, G: 0, B: 0}, // Black
	}
	tableBorders := domain.TableBorders{
		Top:    border,
		Bottom: border,
		Left:   border,
		Right:  border,
	}

	for _, table := range doc.Tables() {
		for _, row := range table.Rows() {
			for _, cell := range row.Cells() {
				_ = cell.SetBorders(tableBorders)
			}
		}
	}

	var buff bytes.Buffer
	if _, err := doc.WriteTo(&buff); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
