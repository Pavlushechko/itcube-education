// internal/repo/catalog_repo.go

package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Pavlushechko/itcube-education/internal/domain"
)

type CatalogRepo struct{ db *pgxpool.Pool }

func NewCatalogRepo(db *pgxpool.Pool) *CatalogRepo { return &CatalogRepo{db: db} }

// -------- Public catalog --------

func (r *CatalogRepo) ListPublishedPrograms(ctx context.Context) ([]domain.Program, error) {
	rows, err := r.db.Query(ctx, `
    select id, title, description, status, created_at
    from programs
    where status='published'
    order by created_at desc
  `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.Program, 0)
	for rows.Next() {
		var p domain.Program
		var st string
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &st, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Status = domain.ProgramStatus(st)
		res = append(res, p)
	}
	return res, rows.Err()
}

// Program page: program + groups (only open groups) for published program
type ProgramWithGroups struct {
	Program domain.Program
	Groups  []domain.Group
}

func (r *CatalogRepo) GetPublishedProgramWithGroups(ctx context.Context, programID uuid.UUID) (ProgramWithGroups, error) {
	row := r.db.QueryRow(ctx, `
		select id, title, description, status, created_at
		from programs
		where id=$1 and status='published'
	`, programID)

	var p domain.Program
	var st string
	if err := row.Scan(&p.ID, &p.Title, &p.Description, &st, &p.CreatedAt); err != nil {
		return ProgramWithGroups{}, err
	}
	p.Status = domain.ProgramStatus(st)

	rows, err := r.db.Query(ctx, `
		select id, program_id, cohort_id, title, capacity, is_open, requires_interview, created_at
		from groups
		where program_id=$1 and is_open=true
		order by created_at desc
	`, programID)
	if err != nil {
		return ProgramWithGroups{}, err
	}
	defer rows.Close()

	gs := make([]domain.Group, 0)
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.ProgramID, &g.CohortID, &g.Title, &g.Capacity, &g.IsOpen, &g.RequiresInterview, &g.CreatedAt); err != nil {
			return ProgramWithGroups{}, err
		}
		gs = append(gs, g)
	}

	return ProgramWithGroups{Program: p, Groups: gs}, rows.Err()
}

// Used by ApplicationService.Create: ensure group open + program published
func (r *CatalogRepo) IsGroupAvailableForApply(ctx context.Context, groupID uuid.UUID) (bool, bool, error) {
	// returns: (programPublished, groupOpen)
	row := r.db.QueryRow(ctx, `
		select p.status, g.is_open
		from groups g
		join programs p on p.id=g.program_id
		where g.id=$1
	`, groupID)
	var pStatus string
	var open bool
	if err := row.Scan(&pStatus, &open); err != nil {
		return false, false, err
	}
	return pStatus == string(domain.ProgramPublished), open, nil
}

func (r *CatalogRepo) GroupRequiresInterview(ctx context.Context, groupID uuid.UUID) (bool, error) {
	row := r.db.QueryRow(ctx, `select requires_interview from groups where id=$1`, groupID)
	var req bool
	return req, row.Scan(&req)
}

func (r *CatalogRepo) IsTeacherInGroup(ctx context.Context, groupID, teacherID uuid.UUID) (bool, error) {
	row := r.db.QueryRow(ctx, `
		select exists(
			select 1 from group_teachers where group_id=$1 and teacher_user_id=$2
		)
	`, groupID, teacherID)
	var ok bool
	return ok, row.Scan(&ok)
}

func (r *CatalogRepo) ListTeacherGroups(ctx context.Context, teacherID uuid.UUID) ([]domain.Group, error) {
	rows, err := r.db.Query(ctx, `
    select g.id, g.program_id, g.cohort_id, g.title, g.capacity, g.is_open, g.requires_interview, g.created_at
    from group_teachers gt
    join groups g on g.id=gt.group_id
    where gt.teacher_user_id=$1
    order by g.created_at desc
  `, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.Group, 0)
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.ProgramID, &g.CohortID, &g.Title, &g.Capacity, &g.IsOpen, &g.RequiresInterview, &g.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, g)
	}
	return res, rows.Err()
}

// -------- Admin CRUD --------

func (r *CatalogRepo) CreateProgramDraft(ctx context.Context, title, desc string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		insert into programs(id, title, description, status)
		values ($1,$2,$3,'draft')
	`, id, title, desc)
	return id, err
}

func (r *CatalogRepo) PublishProgram(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `update programs set status='published' where id=$1`, id)
	return err
}

func (r *CatalogRepo) CreateCohort(ctx context.Context, programID uuid.UUID, year int) (uuid.UUID, error) {
	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		insert into cohorts(id, program_id, year)
		values ($1,$2,$3)
	`, id, programID, year)
	return id, err
}

func (r *CatalogRepo) CreateGroup(ctx context.Context, programID, cohortID uuid.UUID, title string, capacity int, requiresInterview bool, isOpen bool) (uuid.UUID, error) {
	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		insert into groups(id, program_id, cohort_id, title, capacity, requires_interview, is_open)
		values ($1,$2,$3,$4,$5,$6,$7)
	`, id, programID, cohortID, title, capacity, requiresInterview, isOpen)
	return id, err
}

func (r *CatalogRepo) AssignTeacherToGroup(ctx context.Context, groupID, teacherUserID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		insert into group_teachers(group_id, teacher_user_id)
		values ($1,$2)
		on conflict do nothing
	`, groupID, teacherUserID)
	return err
}

func (r *CatalogRepo) SetGroupOpen(ctx context.Context, groupID uuid.UUID, open bool) error {
	_, err := r.db.Exec(ctx, `
		update groups
		set is_open=$2
		where id=$1
	`, groupID, open)
	return err
}

func (r *CatalogRepo) ListAllPrograms(ctx context.Context) ([]domain.Program, error) {
	rows, err := r.db.Query(ctx, `
		select id, title, description, status, created_at
		from programs
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.Program, 0)
	for rows.Next() {
		var p domain.Program
		var st string
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &st, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Status = domain.ProgramStatus(st)
		res = append(res, p)
	}
	return res, rows.Err()
}

func (r *CatalogRepo) GetProgramWithGroupsAdmin(ctx context.Context, programID uuid.UUID) (ProgramWithGroups, error) {
	row := r.db.QueryRow(ctx, `
		select id, title, description, status, created_at
		from programs
		where id=$1
	`, programID)

	var p domain.Program
	var st string
	if err := row.Scan(&p.ID, &p.Title, &p.Description, &st, &p.CreatedAt); err != nil {
		return ProgramWithGroups{}, err
	}
	p.Status = domain.ProgramStatus(st)

	// для staff показываем ВСЕ группы (и закрытые тоже), чтобы админ мог их править
	rows, err := r.db.Query(ctx, `
		select id, program_id, cohort_id, title, capacity, is_open, requires_interview, created_at
		from groups
		where program_id=$1
		order by created_at desc
	`, programID)
	if err != nil {
		return ProgramWithGroups{}, err
	}
	defer rows.Close()

	gs := make([]domain.Group, 0)
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.ProgramID, &g.CohortID, &g.Title, &g.Capacity, &g.IsOpen, &g.RequiresInterview, &g.CreatedAt); err != nil {
			return ProgramWithGroups{}, err
		}
		gs = append(gs, g)
	}

	return ProgramWithGroups{Program: p, Groups: gs}, rows.Err()
}

func (r *CatalogRepo) ListGroupTeachers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		select teacher_user_id
		from group_teachers
		where group_id=$1
		order by teacher_user_id asc
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		res = append(res, id)
	}
	return res, rows.Err()
}

func (r *CatalogRepo) RemoveTeacherFromGroup(ctx context.Context, groupID, teacherUserID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		delete from group_teachers
		where group_id=$1 and teacher_user_id=$2
	`, groupID, teacherUserID)
	return err
}

func (r *CatalogRepo) UpdateGroup(ctx context.Context, groupID uuid.UUID, title *string, capacity *int, isOpen *bool, requiresInterview *bool) error {
	_, err := r.db.Exec(ctx, `
		update groups
		set
			title = coalesce($2, title),
			capacity = coalesce($3, capacity),
			is_open = coalesce($4, is_open),
			requires_interview = coalesce($5, requires_interview)
		where id=$1
	`, groupID, title, capacity, isOpen, requiresInterview)
	return err
}

// Program without "published only" restriction (for staff view)
func (r *CatalogRepo) GetProgram(ctx context.Context, programID uuid.UUID) (domain.Program, error) {
	row := r.db.QueryRow(ctx, `
		select id, title, description, status, created_at
		from programs
		where id=$1
	`, programID)

	var p domain.Program
	var st string
	if err := row.Scan(&p.ID, &p.Title, &p.Description, &st, &p.CreatedAt); err != nil {
		return domain.Program{}, err
	}
	p.Status = domain.ProgramStatus(st)
	return p, nil
}

func (r *CatalogRepo) ListCohortsByProgram(ctx context.Context, programID uuid.UUID) ([]domain.Cohort, error) {
	rows, err := r.db.Query(ctx, `
		select id, program_id, year, created_at
		from cohorts
		where program_id=$1
		order by year desc
	`, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.Cohort, 0)
	for rows.Next() {
		var c domain.Cohort
		if err := rows.Scan(&c.ID, &c.ProgramID, &c.Year, &c.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	return res, rows.Err()
}

func (r *CatalogRepo) ListGroupsByProgram(ctx context.Context, programID uuid.UUID) ([]domain.Group, error) {
	rows, err := r.db.Query(ctx, `
		select id, program_id, cohort_id, title, capacity, is_open, requires_interview, created_at
		from groups
		where program_id=$1
		order by created_at desc
	`, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.Group, 0)
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.ProgramID, &g.CohortID, &g.Title, &g.Capacity, &g.IsOpen, &g.RequiresInterview, &g.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, g)
	}
	return res, rows.Err()
}

// teacher = назначение, не роль
func (r *CatalogRepo) ListTeacherGroupsByProgram(ctx context.Context, teacherID, programID uuid.UUID) ([]domain.Group, error) {
	rows, err := r.db.Query(ctx, `
		select g.id, g.program_id, g.cohort_id, g.title, g.capacity, g.is_open, g.requires_interview, g.created_at
		from group_teachers gt
		join groups g on g.id = gt.group_id
		where gt.teacher_user_id=$1 and g.program_id=$2
		order by g.created_at desc
	`, teacherID, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.Group, 0)
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.ProgramID, &g.CohortID, &g.Title, &g.Capacity, &g.IsOpen, &g.RequiresInterview, &g.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, g)
	}
	return res, rows.Err()
}

func (r *CatalogRepo) UpdateProgram(ctx context.Context, id uuid.UUID, title *string, desc *string) error {
	_, err := r.db.Exec(ctx, `
		update programs
		set
			title = coalesce($2, title),
			description = coalesce($3, description)
		where id=$1
	`, id, title, desc)
	return err
}

func (r *CatalogRepo) GetCohortByProgramYear(ctx context.Context, programID uuid.UUID, year int) (domain.Cohort, bool, error) {
	row := r.db.QueryRow(ctx, `
		select id, program_id, year, created_at
		from cohorts
		where program_id=$1 and year=$2
	`, programID, year)

	var c domain.Cohort
	err := row.Scan(&c.ID, &c.ProgramID, &c.Year, &c.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Cohort{}, false, nil
		}
		return domain.Cohort{}, false, err
	}
	return c, true, nil
}

// true если teacher назначен хотя бы на одну группу этой программы
func (r *CatalogRepo) IsTeacherInProgram(ctx context.Context, teacherID, programID uuid.UUID) (bool, error) {
	row := r.db.QueryRow(ctx, `
		select exists(
			select 1
			from group_teachers gt
			join groups g on g.id = gt.group_id
			where gt.teacher_user_id=$1 and g.program_id=$2
		)
	`, teacherID, programID)

	var ok bool
	return ok, row.Scan(&ok)
}

func (r *CatalogRepo) GetGroupProgramID(ctx context.Context, groupID uuid.UUID) (uuid.UUID, error) {
	row := r.db.QueryRow(ctx, `select program_id from groups where id=$1`, groupID)
	var pid uuid.UUID
	return pid, row.Scan(&pid)
}
