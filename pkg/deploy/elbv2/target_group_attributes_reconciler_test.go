package elbv2

import (
	"context"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	elbv2sdk "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/aws-load-balancer-controller/pkg/aws/services"
	elbv2model "sigs.k8s.io/aws-load-balancer-controller/pkg/model/elbv2"
)

func Test_defaultTargetGroupAttributeReconciler_getDesiredTargetGroupAttributes(t *testing.T) {
	type args struct {
		resTG *elbv2model.TargetGroup
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "standard case",
			args: args{
				resTG: &elbv2model.TargetGroup{
					Spec: elbv2model.TargetGroupSpec{
						TargetGroupAttributes: []elbv2model.TargetGroupAttribute{
							{
								Key:   "keyA",
								Value: "valueA",
							},
							{
								Key:   "keyB",
								Value: "valueB",
							},
						},
					},
				},
			},
			want: map[string]string{
				"keyA": "valueA",
				"keyB": "valueB",
			},
		},
		{
			name: "nil attributes case",
			args: args{
				resTG: &elbv2model.TargetGroup{
					Spec: elbv2model.TargetGroupSpec{
						TargetGroupAttributes: nil,
					},
				},
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &defaultTargetGroupAttributeReconciler{}
			got := r.getDesiredTargetGroupAttributes(context.Background(), tt.args.resTG)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_defaultTargetGroupAttributeReconciler_getCurrentTargetGroupAttributes(t *testing.T) {
	type describeTargetGroupAttributesWithContextCall struct {
		req  *elbv2sdk.DescribeTargetGroupAttributesInput
		resp *elbv2sdk.DescribeTargetGroupAttributesOutput
		err  error
	}
	type fields struct {
		describeTargetGroupAttributesWithContextCalls []describeTargetGroupAttributesWithContextCall
	}
	type args struct {
		sdkTG TargetGroupWithTags
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr error
	}{
		{
			name: "standard case",
			fields: fields{
				describeTargetGroupAttributesWithContextCalls: []describeTargetGroupAttributesWithContextCall{
					{
						req: &elbv2sdk.DescribeTargetGroupAttributesInput{
							TargetGroupArn: awssdk.String("my-arn"),
						},
						resp: &elbv2sdk.DescribeTargetGroupAttributesOutput{
							Attributes: []*elbv2sdk.TargetGroupAttribute{
								{
									Key:   awssdk.String("keyA"),
									Value: awssdk.String("valueA"),
								},
								{
									Key:   awssdk.String("keyB"),
									Value: awssdk.String("valueB"),
								},
							},
						},
					},
				},
			},
			args: args{
				sdkTG: TargetGroupWithTags{
					TargetGroup: &elbv2sdk.TargetGroup{
						TargetGroupArn: awssdk.String("my-arn"),
					},
					Tags: nil,
				},
			},
			want: map[string]string{
				"keyA": "valueA",
				"keyB": "valueB",
			},
		},
		{
			name: "error case",
			fields: fields{
				describeTargetGroupAttributesWithContextCalls: []describeTargetGroupAttributesWithContextCall{
					{
						req: &elbv2sdk.DescribeTargetGroupAttributesInput{
							TargetGroupArn: awssdk.String("my-arn"),
						},
						err: errors.New("some error"),
					},
				},
			},
			args: args{
				sdkTG: TargetGroupWithTags{
					TargetGroup: &elbv2sdk.TargetGroup{
						TargetGroupArn: awssdk.String("my-arn"),
					},
					Tags: nil,
				},
			},
			wantErr: errors.New("some error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			elbv2Client := services.NewMockELBV2(ctrl)
			for _, call := range tt.fields.describeTargetGroupAttributesWithContextCalls {
				elbv2Client.EXPECT().DescribeTargetGroupAttributesWithContext(gomock.Any(), call.req).Return(call.resp, call.err)
			}

			r := &defaultTargetGroupAttributeReconciler{
				elbv2Client: elbv2Client,
			}
			got, err := r.getCurrentTargetGroupAttributes(context.Background(), tt.args.sdkTG)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
