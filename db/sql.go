package db

const (
	addUserSQL = `INSERT INTO user(username, nickname, is_admin, is_enable, id_card, phone, qq, t_shirt)
 VALUES(:username, :nickname, :is_admin, :is_enable, :id_card, :phone, :qq, :t_shirt)`
	updUserSQL       = "UPDATE user SET nickname=:nickname, id_card=:id_card, phone=:phone, qq=:qq, t_shirt=:t_shirt WHERE username=:username"
	updUserEnableSQL = "UPDATE user SET is_enable=:is_enable WHERE username=:username"
	updUserAdminSQL  = "UPDATE user SET is_admin=:is_admin WHERE username=:username"

	addTeamSQL        = "INSERT INTO team(name, is_enable, is_self) VALUES(:name, :is_enable, :is_self)"
	addTeamUserRelSQL = "INSERT INTO team_user_rel(team_id, username) VALUES(:team_id, :username)"
	updTeamSQL        = "UPDATE team SET name=:name WHERE id=:id"
	updTeamEnableSQL  = "UPDATE team SET is_enable=:is_enable WHERE id=:id"

	addContestProblemSQL = "INSERT INTO contest_problem(contest_id, oj_id, pid, `index`) VALUES(:contest_id, :oj_id, :pid, :index)"
)
