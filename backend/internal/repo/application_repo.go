// internal/repo/application_repo.go

package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"strconv"

	"github.com/Pavlushechko/itcube-education/internal/domain"
)

type ApplicationRepo struct {
	db *pgxpool.Pool
}

type ApplicationView struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	GroupID   uuid.UUID
	Status    string
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time

	ProgramID    uuid.UUID
	ProgramTitle string
	GroupTitle   string

	InterviewResult  *string
	InterviewComment *string
	InterviewByRole  *string
	InterviewAt      *time.Time
}

func NewApplicationRepo(db *pgxpool.Pool) *ApplicationRepo {
	return &ApplicationRepo{db: db}
}

func (r *ApplicationRepo) Create(ctx context.Context, a domain.EnrollmentApplication) error {
	_, err := r.db.Exec(ctx, `
		insert into enrollment_applications(id, user_id, group_id, status, comment, created_at, updated_at)
		values ($1,$2,$3,$4,$5, now(), now())
	`, a.ID, a.UserID, a.GroupID, a.Status, a.Comment)
	return err
}

func (r *ApplicationRepo) Get(ctx context.Context, id uuid.UUID) (domain.EnrollmentApplication, error) {
	row := r.db.QueryRow(ctx, `
		select id, user_id, group_id, status, comment, created_at, updated_at
		from enrollment_applications
		where id=$1
	`, id)

	var a domain.EnrollmentApplication
	var status string
	err := row.Scan(&a.ID, &a.UserID, &a.GroupID, &status, &a.Comment, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return domain.EnrollmentApplication{}, err
	}
	a.Status = domain.ApplicationStatus(status)
	return a, nil
}

func (r *ApplicationRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.EnrollmentApplication, error) {
	rows, err := r.db.Query(ctx, `
		select id, user_id, group_id, status, comment, created_at, updated_at
		from enrollment_applications
		where user_id=$1
		order by created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.EnrollmentApplication
	for rows.Next() {
		var a domain.EnrollmentApplication
		var status string
		if err := rows.Scan(&a.ID, &a.UserID, &a.GroupID, &status, &a.Comment, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Status = domain.ApplicationStatus(status)
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *ApplicationRepo) ListByFilter(ctx context.Context, groupID *uuid.UUID, status *string) ([]domain.EnrollmentApplication, error) {
	// простая динамика без билдера
	q := `
		select id, user_id, group_id, status, comment, created_at, updated_at
		from enrollment_applications
		where 1=1
	`
	args := []any{}
	i := 1
	if groupID != nil {
		q += " and group_id=$" + itoa(i)
		args = append(args, *groupID)
		i++
	}
	if status != nil {
		q += " and status=$" + itoa(i)
		args = append(args, *status)
		i++
	}
	q += " order by created_at desc"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.EnrollmentApplication
	for rows.Next() {
		var a domain.EnrollmentApplication
		var st string
		if err := rows.Scan(&a.ID, &a.UserID, &a.GroupID, &st, &a.Comment, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Status = domain.ApplicationStatus(st)
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *ApplicationRepo) ListByFilterView(ctx context.Context, groupID *uuid.UUID, status *string) ([]ApplicationView, error) {
	q := `
		select a.id, a.user_id, a.group_id, a.status, a.comment, a.created_at, a.updated_at,
		       p.id as program_id, p.title as program_title, g.title as group_title,
		       i.result, i.comment, i.interviewer_role, i.updated_at
		from enrollment_applications a
		join groups g on g.id = a.group_id
  		join programs p on p.id = g.program_id
		left join interviews i on i.application_id = a.id
		where 1=1
	`
	args := []any{}
	i := 1

	if groupID != nil {
		q += " and a.group_id=$" + strconv.Itoa(i)
		args = append(args, *groupID)
		i++
	}
	if status != nil {
		q += " and a.status=$" + strconv.Itoa(i)
		args = append(args, *status)
		i++
	}

	q += " order by a.created_at desc"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]ApplicationView, 0)
	for rows.Next() {
		var a ApplicationView
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.GroupID, &a.Status, &a.Comment, &a.CreatedAt, &a.UpdatedAt,
			&a.ProgramID, &a.ProgramTitle, &a.GroupTitle,
			&a.InterviewResult, &a.InterviewComment, &a.InterviewByRole, &a.InterviewAt,
		); err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *ApplicationRepo) ListByProgramView(ctx context.Context, programID uuid.UUID, status *string) ([]ApplicationView, error) {
	q := `
		select a.id, a.user_id, a.group_id, a.status, a.comment, a.created_at, a.updated_at,
		       p.id as program_id, p.title as program_title, g.title as group_title,
		       i.result, i.comment, i.interviewer_role, i.updated_at
		from enrollment_applications a
		join groups g on g.id = a.group_id
  		join programs p on p.id = g.program_id
		left join interviews i on i.application_id = a.id
		where g.program_id = $1
	`
	args := []any{programID}
	i := 2

	if status != nil {
		q += " and a.status=$" + strconv.Itoa(i)
		args = append(args, *status)
		i++
	}
	q += " order by a.created_at desc"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]ApplicationView, 0)
	for rows.Next() {
		var a ApplicationView
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.GroupID, &a.Status, &a.Comment, &a.CreatedAt, &a.UpdatedAt,
			&a.ProgramID, &a.ProgramTitle, &a.GroupTitle,
			&a.InterviewResult, &a.InterviewComment, &a.InterviewByRole, &a.InterviewAt,
		); err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *ApplicationRepo) ListForTeacherByProgramView(ctx context.Context, teacherID, programID uuid.UUID, status *string) ([]ApplicationView, error) {
	q := `
		select a.id, a.user_id, a.group_id, a.status, a.comment, a.created_at, a.updated_at,
		       p.id as program_id, p.title as program_title, g.title as group_title,
		       i.result, i.comment, i.interviewer_role, i.updated_at
		from enrollment_applications a
		join groups g on g.id = a.group_id
  		join programs p on p.id = g.program_id
		join group_teachers gt on gt.group_id = g.id
		left join interviews i on i.application_id = a.id
		where gt.teacher_user_id = $1 and g.program_id = $2
	`
	args := []any{teacherID, programID}
	i := 3

	if status != nil {
		q += " and a.status=$" + strconv.Itoa(i)
		args = append(args, *status)
		i++
	}
	q += " order by a.created_at desc"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]ApplicationView, 0)
	for rows.Next() {
		var a ApplicationView
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.GroupID, &a.Status, &a.Comment, &a.CreatedAt, &a.UpdatedAt,
			&a.ProgramID, &a.ProgramTitle, &a.GroupTitle,
			&a.InterviewResult, &a.InterviewComment, &a.InterviewByRole, &a.InterviewAt,
		); err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *ApplicationRepo) ListAllView(ctx context.Context, status *string) ([]ApplicationView, error) {
	q := `
		select a.id, a.user_id, a.group_id, a.status, a.comment, a.created_at, a.updated_at,
		       p.id as program_id, p.title as program_title, g.title as group_title,
		       i.result, i.comment, i.interviewer_role, i.updated_at
		from enrollment_applications a
	  	join groups g on g.id = a.group_id
  		join programs p on p.id = g.program_id
		left join interviews i on i.application_id = a.id
		where 1=1
	`
	args := []any{}
	i := 1

	if status != nil {
		q += " and a.status=$" + strconv.Itoa(i)
		args = append(args, *status)
		i++
	}
	q += " order by a.created_at desc"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]ApplicationView, 0)
	for rows.Next() {
		var a ApplicationView
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.GroupID, &a.Status, &a.Comment, &a.CreatedAt, &a.UpdatedAt,
			&a.ProgramID, &a.ProgramTitle, &a.GroupTitle,
			&a.InterviewResult, &a.InterviewComment, &a.InterviewByRole, &a.InterviewAt,
		); err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *ApplicationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, to domain.ApplicationStatus) error {
	_, err := r.db.Exec(ctx, `
		update enrollment_applications
		set status=$2, updated_at=now()
		where id=$1
	`, id, string(to))
	return err
}

func (r *ApplicationRepo) InsertAudit(ctx context.Context, appID uuid.UUID, actorID uuid.UUID, actorRole string, from, to domain.ApplicationStatus, reason string) error {
	_, err := r.db.Exec(ctx, `
		insert into application_status_audit(id, application_id, actor_user_id, actor_role, from_status, to_status, reason, created_at)
		values ($1,$2,$3,$4,$5,$6,$7, now())
	`, uuid.New(), appID, actorID, actorRole, string(from), string(to), reason)
	return err
}

func (r *ApplicationRepo) CountEnrollmentsByGroup(ctx context.Context, groupID uuid.UUID) (int, error) {
	row := r.db.QueryRow(ctx, `select count(*) from enrollments where group_id=$1`, groupID)
	var n int
	return n, row.Scan(&n)
}

func (r *ApplicationRepo) GroupCapacity(ctx context.Context, groupID uuid.UUID) (int, error) {
	row := r.db.QueryRow(ctx, `select capacity from groups where id=$1`, groupID)
	var cap int
	return cap, row.Scan(&cap)
}

func (r *ApplicationRepo) CreateEnrollment(ctx context.Context, userID, groupID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		insert into enrollments(id, user_id, group_id, created_at)
		values ($1,$2,$3, now())
		on conflict (user_id, group_id) do nothing
	`, uuid.New(), userID, groupID)
	return err
}

func (r *ApplicationRepo) HasEnrollment(ctx context.Context, userID, groupID uuid.UUID) (bool, error) {
	row := r.db.QueryRow(ctx, `
		select exists(select 1 from enrollments where user_id=$1 and group_id=$2)
	`, userID, groupID)
	var ok bool
	return ok, row.Scan(&ok)
}

func (r *ApplicationRepo) ListEnrolledUsersByGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		select user_id
		from enrollments
		where group_id=$1
		order by created_at asc
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []uuid.UUID
	for rows.Next() {
		var uid uuid.UUID
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		res = append(res, uid)
	}
	return res, rows.Err()
}

func (r *ApplicationRepo) ListByProgram(ctx context.Context, programID uuid.UUID, status *string) ([]domain.EnrollmentApplication, error) {
	q := `
		select a.id, a.user_id, a.group_id, a.status, a.comment, a.created_at, a.updated_at
		from enrollment_applications a
		join groups g on g.id = a.group_id
		where g.program_id = $1
	`
	args := []any{programID}
	i := 2

	if status != nil {
		q += " and a.status=$" + strconv.Itoa(i)
		args = append(args, *status)
		i++
	}
	q += " order by a.created_at desc"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.EnrollmentApplication, 0)
	for rows.Next() {
		var a domain.EnrollmentApplication
		var st string
		if err := rows.Scan(&a.ID, &a.UserID, &a.GroupID, &st, &a.Comment, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Status = domain.ApplicationStatus(st)
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *ApplicationRepo) ListForTeacherByProgram(ctx context.Context, teacherID, programID uuid.UUID, status *string) ([]domain.EnrollmentApplication, error) {
	q := `
		select a.id, a.user_id, a.group_id, a.status, a.comment, a.created_at, a.updated_at
		from enrollment_applications a
		join groups g on g.id = a.group_id
		join group_teachers gt on gt.group_id = g.id
		where gt.teacher_user_id = $1 and g.program_id = $2
	`
	args := []any{teacherID, programID}
	i := 3

	if status != nil {
		q += " and a.status=$" + strconv.Itoa(i)
		args = append(args, *status)
		i++
	}
	q += " order by a.created_at desc"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.EnrollmentApplication, 0)
	for rows.Next() {
		var a domain.EnrollmentApplication
		var st string
		if err := rows.Scan(&a.ID, &a.UserID, &a.GroupID, &st, &a.Comment, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Status = domain.ApplicationStatus(st)
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *ApplicationRepo) ListAll(ctx context.Context, status *string) ([]domain.EnrollmentApplication, error) {
	q := `
		select id, user_id, group_id, status, comment, created_at, updated_at
		from enrollment_applications
		where 1=1
	`
	args := []any{}
	i := 1

	if status != nil {
		q += " and status=$" + strconv.Itoa(i)
		args = append(args, *status)
		i++
	}

	q += " order by created_at desc"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.EnrollmentApplication, 0)
	for rows.Next() {
		var a domain.EnrollmentApplication
		var st string
		if err := rows.Scan(&a.ID, &a.UserID, &a.GroupID, &st, &a.Comment, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Status = domain.ApplicationStatus(st)
		res = append(res, a)
	}
	return res, rows.Err()
}

func itoa(i int) string { // чтобы не тащить strconv в каждый файл
	return string(rune('0' + i))
}

// NOTE: itoa выше годится только до 9 параметров; для MVP ок.
// Потом заменишь на strconv.Itoa(i).
var _ = time.Now
