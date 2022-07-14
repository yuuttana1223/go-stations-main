package sta12_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mattn/go-sqlite3"

	"github.com/TechBowl-japan/go-stations/db"
	"github.com/TechBowl-japan/go-stations/model"
	"github.com/TechBowl-japan/go-stations/service"
)

func TestStation12(t *testing.T) {
	t.Parallel()

	dbpath := "./temp_test.db"
	d, err := db.NewDB(dbpath)
	if err != nil {
		t.Errorf("データベースの作成に失敗しました: %v", err)
		return
	}

	t.Cleanup(func() {
		if err := d.Close(); err != nil {
			t.Errorf("データベースのクローズに失敗しました: %v", err)
			return
		}
		if err := os.Remove(dbpath); err != nil {
			t.Errorf("テスト用のDBファイルの削除に失敗しました: %v", err)
			return
		}
	})

	stmt, err := d.Prepare(`INSERT INTO todos(subject) VALUES(?)`)
	if err != nil {
		t.Errorf("ステートメントの作成に失敗しました: %v", err)
		return
	}
	t.Cleanup(func() {
		if err := stmt.Close(); err != nil {
			t.Errorf("ステートメントのクローズに失敗しました: %v", err)
			return
		}
	})

	_, err = stmt.Exec("todo subject")
	if err != nil {
		t.Errorf("todoの追加に失敗しました: %v", err)
		return
	}

	testcases := map[string]struct {
		ID          int64
		Subject     string
		Description string
		WantError   error
	}{
		"ID is empty": {
			WantError: &model.ErrNotFound{},
		},
		"Subject is empty": {
			ID:        1,
			WantError: sqlite3.ErrConstraint,
		},
		"Description is empty": {
			ID:      1,
			Subject: "todo subject 1",
		},
		"Have not empty arguments": {
			ID:          1,
			Subject:     "todo subject 2",
			Description: "todo description 2",
		},
	}

	for name, tc := range testcases {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			svc := service.NewTODOService(d)
			got, err := svc.UpdateTODO(context.Background(), tc.ID, tc.Subject, tc.Description)
			switch tc.WantError {
			case nil:
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
					return
				}
			default:
				if !errors.As(err, &tc.WantError) {
					t.Errorf("期待していないエラーの Type です, got = %t, want = %+v", err, tc.WantError)
					return
				}
				return
			}

			want := &model.TODO{
				ID:          tc.ID,
				Subject:     tc.Subject,
				Description: tc.Description,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if diff := cmp.Diff(got, want, cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Error("期待していない値です\n", diff)
			}
		})
	}
}
