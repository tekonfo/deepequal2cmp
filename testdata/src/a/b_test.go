package a

import (
	"page-re-player-app-server/app/model"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetHeatmapDetailResponse(t *testing.T) {
	type args struct {
		heatmapDetail model.Heatmap
		heatmapLog    model.HeatmapLog
	}
	tests := []struct {
		name string
		args args
		want HeatmapDetailResponse
	}{
		{
			name: "simple test",
			args: args{
				heatmapDetail: model.Heatmap{
					Model: model.Model{
						ID: 1,
					},
					DomainID: 1,
					Path:     "hogehoge",
				},
				heatmapLog: model.HeatmapLog{},
			},
			want: HeatmapDetailResponse{
				Data: HeatmapDetailFormat{
					Id:      uint(1),
					HtmlUrl: "/heatmap-page/1/hogehoge/index.html",
					Pc: ViewTypeInfoFormat{
						Width:  PCWidth,
						Height: PCHeight,
					},
					Sp: ViewTypeInfoFormat{
						Width:  SPWidth,
						Height: SPHeight,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetHeatmapDetailResponse(tt.args.heatmapDetail, tt.args.heatmapLog)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("GetHeatmapLog() = (-got +want)\n%s", diff)
			}
		})
	}
}
