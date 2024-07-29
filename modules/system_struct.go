package modules

type Percpu struct {
	CPU       string
	User      float64
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	Nice      float64 `json:"nice"`
	Iowait    float64 `json:"iowait"`
	Irq       float64 `json:"irq"`
	Softirq   float64 `json:"softirq"`
	Steal     float64 `json:"steal"`
	Guest     float64 `json:"guest"`
	GuestNice float64 `json:"guestNice"`
}

type Totalcpu struct {
	CPU       string
	User      float64
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	Nice      float64 `json:"nice"`
	Iowait    float64 `json:"iowait"`
	Irq       float64 `json:"irq"`
	Softirq   float64 `json:"softirq"`
	Steal     float64 `json:"steal"`
	Guest     float64 `json:"guest"`
	GuestNice float64 `json:"guestNice"`
}

type Memory struct {
	// Total amount of RAM on this system
	Total uint64
	// RAM available for programs to allocate
	Available uint64
	// RAM used by programs
	Used uint64
	// Percentage of RAM used by programs
	UsedPercent float64
}
