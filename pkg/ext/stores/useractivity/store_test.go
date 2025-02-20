package useractivity

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	ext "github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1"
	v3Legacy "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	wranglerfake "github.com/rancher/wrangler/v3/pkg/generic/fake"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestStore_create(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockTokenControllerFake := wranglerfake.NewMockNonNamespacedControllerInterface[*v3Legacy.Token, *v3Legacy.TokenList](ctrl)
	mockTokenCacheFake := wranglerfake.NewMockNonNamespacedCacheInterface[*v3Legacy.Token](ctrl)
	mockUserCacheFake := wranglerfake.NewMockNonNamespacedCacheInterface[*v3.User](ctrl)
	uas := &Store{
		tokens:     mockTokenControllerFake,
		tokenCache: mockTokenCacheFake,
		userCache:  mockUserCacheFake,
	}

	type args struct {
		in0          context.Context
		userActivity *ext.UserActivity
		token        *v3Legacy.Token
		lastActivity v1.Time
		idleMins     int
		dryRun       bool
	}
	tests := []struct {
		name      string
		args      args
		mockSetup func()
		want      *ext.UserActivity
		wantErr   bool
	}{
		{
			name: "valid useractivity is created",
			args: args{
				in0: nil,
				userActivity: &ext.UserActivity{
					ObjectMeta: v1.ObjectMeta{
						Name: "u-mo773yttt4",
					},
				},
				token: &v3Legacy.Token{
					ObjectMeta: v1.ObjectMeta{
						Name: "u-mo773yttt4",
					},
					UserID: "u-mo773yttt4",
				},
				lastActivity: v1.Time{
					Time: time.Date(2025, 1, 31, 16, 44, 0, 0, &time.Location{}),
				},
				idleMins: 10,
			},
			mockSetup: func() {
				// we don't care about the object returned by the Patch function,
				// since we only check there are no errors.
				mockTokenControllerFake.EXPECT().Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(name string, pt types.PatchType, data []byte, subresources ...string) (*v3Legacy.Token, error) {
						return &v3Legacy.Token{}, nil
					},
				).Times(1)
			},
			want: &ext.UserActivity{
				ObjectMeta: v1.ObjectMeta{
					Name: "u-mo773yttt4",
				},
				Status: ext.UserActivityStatus{
					ExpiresAt: time.Date(2025, 1, 31, 16, 54, 0, 0, &time.Location{}).String(),
				},
			},
			wantErr: false,
		},
		{
			name: "error updating token value LastIdleTimeout",
			args: args{
				in0: nil,
				userActivity: &ext.UserActivity{
					ObjectMeta: v1.ObjectMeta{
						Name: "u-mo773yttt4",
					},
				},
				token: &v3Legacy.Token{
					ObjectMeta: v1.ObjectMeta{},
					UserID:     "u-mo773yttt4",
				},
				lastActivity: v1.Time{
					Time: time.Date(2025, 1, 31, 16, 44, 0, 0, &time.Location{}).UTC(),
				},
				idleMins: 10,
			},
			mockSetup: func() {
				mockTokenControllerFake.EXPECT().Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(name string, pt types.PatchType, data []byte, subresources ...string) (*v3Legacy.Token, error) {
						return nil, errors.New("some error happend")
					},
				).AnyTimes()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "create useractivity with dry-run option",
			args: args{
				in0: nil,
				userActivity: &ext.UserActivity{
					ObjectMeta: v1.ObjectMeta{
						Name: "u-mo773yttt4",
					},
				},
				token: &v3Legacy.Token{
					ObjectMeta: v1.ObjectMeta{
						Name: "u-mo773yttt4",
					},
					UserID: "u-mo773yttt4",
				},
				lastActivity: v1.Time{
					Time: time.Date(2025, 1, 31, 16, 44, 0, 0, &time.Location{}),
				},
				idleMins: 10,
				dryRun:   true,
			},
			mockSetup: func() {},
			want: &ext.UserActivity{
				ObjectMeta: v1.ObjectMeta{
					Name: "u-mo773yttt4",
				},
				Status: ext.UserActivityStatus{
					ExpiresAt: time.Date(2025, 1, 31, 16, 54, 0, 0, &time.Location{}).String(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			got, err := uas.create(tt.args.in0, tt.args.userActivity, tt.args.token, tt.args.lastActivity, tt.args.idleMins, tt.args.dryRun)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Store.create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_get(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockTokenControllerFake := wranglerfake.NewMockNonNamespacedControllerInterface[*v3Legacy.Token, *v3Legacy.TokenList](ctrl)
	mockUserCacheFake := wranglerfake.NewMockNonNamespacedCacheInterface[*v3.User](ctrl)
	uas := &Store{
		tokens:     mockTokenControllerFake,
		tokenCache: mockTokenControllerFake.Cache(),
		userCache:  mockUserCacheFake,
	}
	contextBG := context.Background()
	type args struct {
		ctx     context.Context
		name    string
		options *v1.GetOptions
	}
	tests := []struct {
		name      string
		args      args
		mockSetup func()
		want      runtime.Object
		wantErr   bool
	}{
		{
			name: "valid useractivity retrieved",
			args: args{
				ctx:     contextBG,
				name:    "ua_admin_token-12345",
				options: &v1.GetOptions{},
			},
			mockSetup: func() {
				mockTokenControllerFake.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&v3Legacy.Token{
					ObjectMeta: v1.ObjectMeta{
						Name: "token-12345",
					},
					UserID: "admin",
					ActivityLastSeenAt: &v1.Time{
						Time: time.Date(2025, 1, 31, 16, 44, 0, 0, &time.Location{}),
					},
				}, nil).Times(1)
			},
			want: &ext.UserActivity{
				ObjectMeta: v1.ObjectMeta{
					Name: "token-12345",
				},
				Status: ext.UserActivityStatus{
					ExpiresAt: time.Date(2025, 1, 31, 16, 44, 0, 0, &time.Location{}).String(),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid useractivity name",
			args: args{
				ctx:     contextBG,
				name:    "ua_admin_token_12345",
				options: &v1.GetOptions{},
			},
			mockSetup: func() {},
			want:      nil,
			wantErr:   true,
		},
		{
			name: "invalid token retrieved",
			args: args{
				ctx:     contextBG,
				name:    "ua_admin_token-12345",
				options: &v1.GetOptions{},
			},
			mockSetup: func() {
				mockTokenControllerFake.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("invalid token name")).Times(1)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid user name retrieved",
			args: args{
				ctx:     contextBG,
				name:    "ua_user1_token-12345",
				options: &v1.GetOptions{},
			},
			mockSetup: func() {
				mockTokenControllerFake.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&v3Legacy.Token{
					UserID: "token-12345",
					ActivityLastSeenAt: &v1.Time{
						Time: time.Date(2025, 1, 31, 16, 44, 0, 0, &time.Location{}),
					},
				}, nil).Times(1)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt.mockSetup()
		t.Run(tt.name, func(t *testing.T) {
			got, err := uas.get(tt.args.ctx, tt.args.name, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Store.get() = %v, want %v", got, tt.want)
			}
		})
	}
}
