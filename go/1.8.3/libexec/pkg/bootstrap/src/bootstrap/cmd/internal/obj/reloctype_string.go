// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/internal/obj/reloctype_string.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/internal/obj/reloctype_string.go:1
// Code generated by "stringer -type=RelocType"; DO NOT EDIT

package obj

import "fmt"

const _RelocType_name = "R_ADDRR_ADDRPOWERR_ADDRARM64R_ADDRMIPSR_ADDROFFR_WEAKADDROFFR_SIZER_CALLR_CALLARMR_CALLARM64R_CALLINDR_CALLPOWERR_CALLMIPSR_CONSTR_PCRELR_TLS_LER_TLS_IER_GOTOFFR_PLT0R_PLT1R_PLT2R_USEFIELDR_USETYPER_METHODOFFR_POWER_TOCR_GOTPCRELR_JMPMIPSR_DWARFREFR_ARM64_TLS_LER_ARM64_TLS_IER_ARM64_GOTPCRELR_POWER_TLS_LER_POWER_TLS_IER_POWER_TLSR_ADDRPOWER_DSR_ADDRPOWER_GOTR_ADDRPOWER_PCRELR_ADDRPOWER_TOCRELR_ADDRPOWER_TOCREL_DSR_PCRELDBLR_ADDRMIPSUR_ADDRMIPSTLS"

var _RelocType_index = [...]uint16{0, 6, 17, 28, 38, 47, 60, 66, 72, 81, 92, 101, 112, 122, 129, 136, 144, 152, 160, 166, 172, 178, 188, 197, 208, 219, 229, 238, 248, 262, 276, 292, 306, 320, 331, 345, 360, 377, 395, 416, 426, 437, 450}

func (i RelocType) String() string {
	i -= 1
	if i < 0 || i >= RelocType(len(_RelocType_index)-1) {
		return fmt.Sprintf("RelocType(%d)", i+1)
	}
	return _RelocType_name[_RelocType_index[i]:_RelocType_index[i+1]]
}
