`go run .\cmd\api`

`
# Headers
$admin = @{ "X-User-Id"="aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa"; "X-Role"="admin" }
$user  = @{ "X-User-Id"="11111111-1111-1111-1111-111111111111"; "X-Role"="user" }
$mod   = @{ "X-User-Id"="33333333-3333-3333-3333-333333333333"; "X-Role"="moderator" }
$teacher = @{ "X-User-Id"="22222222-2222-2222-2222-222222222222"; "X-Role"="teacher" }

# 1) Create program
$p = Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/admin/programs" -Headers $admin -ContentType "application/json" `
-Body (@{title="Go backend"; description="описание"} | ConvertTo-Json)

# 2) Create cohort
$c = Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/admin/cohorts" -Headers $admin -ContentType "application/json" `
-Body (@{program_id=$p.id; year=2026} | ConvertTo-Json)

# 3) Create group
$g = Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/admin/groups" -Headers $admin -ContentType "application/json" `
-Body (@{program_id=$p.id; cohort_id=$c.id; title="Group A"; capacity=30; requires_interview=$true; is_open=$true} | ConvertTo-Json)
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/admin/groups/$($g.id)/teachers?teacher_user_id=$($teacher.'X-User-Id')" -Headers $admin

# 4) Publish program
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/admin/programs/$($p.id)/publish" -Headers $admin

# 5) User creates application
$app = Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/enrollments/applications" -Headers $user -ContentType "application/json" `
-Body (@{group_id=$g.id; comment="хочу учиться"} | ConvertTo-Json)

# 6) Moderator moves to in_review
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/admin/applications/$($app.id)/status" -Headers $mod -ContentType "application/json" `
-Body (@{status="in_review"; reason="взято"} | ConvertTo-Json)

Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/teacher/applications/$($app.id)/interview" -Headers $teacher -ContentType "application/json" `
-Body (@{result="recommended"; comment="ок"} | ConvertTo-Json)

Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/admin/applications/$($app.id)/status" -Headers $mod -ContentType "application/json" `
-Body (@{status="approved"; reason="принят"} | ConvertTo-Json)

Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/teacher/groups/$($g.id)/materials" -Headers $teacher -ContentType "application/json" `
-Body (@{type="link"; title="Вводная лекция"; content="https://example.com/intro"} | ConvertTo-Json)

Invoke-RestMethod -Method Get -Uri "http://127.0.0.1:8080/learn/groups/$($g.id)/materials" -Headers $user


"OK. program=$($p.id) cohort=$($c.id) group=$($g.id) app=$($app.id)"

`