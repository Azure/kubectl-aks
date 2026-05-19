package check

import "testing"

func TestIGCommandGeneration(t *testing.T) {
	tests := []struct {
		name string
		ig   IGCheck
		want string
	}{
		{
			name: "dns trace",
			ig:   IGCheck{GadgetImage: "trace_dns", Filters: []string{"rcode!=Success", "qr=R"}},
			want: "ig run trace_dns --host --timeout {{.Duration}} --output json --filter rcode!=Success,qr=R 2>/dev/null || true",
		},
		{
			name: "tcp drops",
			ig:   IGCheck{GadgetImage: "trace_tcpretrans", Filters: []string{"type=LOSS"}},
			want: "ig run trace_tcpretrans --host --timeout {{.Duration}} --output json --filter type=LOSS 2>/dev/null || true",
		},
		{
			name: "custom output mode",
			ig:   IGCheck{GadgetImage: "trace_dns", OutputMode: "columns"},
			want: "ig run trace_dns --host --timeout {{.Duration}} --output columns 2>/dev/null || true",
		},
		{
			name: "with extra args",
			ig:   IGCheck{GadgetImage: "trace_dns", Filters: []string{"qr=R"}, ExtraArgs: []string{"--fields", "name,rcode"}},
			want: "ig run trace_dns --host --timeout {{.Duration}} --output json --filter qr=R --fields name,rcode 2>/dev/null || true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ig.IGCommand()
			if got != tt.want {
				t.Errorf("IGCommand() =\n  %q\nwant:\n  %q", got, tt.want)
			}
		})
	}
}
