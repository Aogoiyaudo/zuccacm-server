package db

const (
	addUserSQL = `INSERT INTO user(username, nickname, is_admin, is_enable, id_card, phone, qq, t_shirt)
 VALUES(:username, :nickname, :is_admin, :is_enable, :id_card, :phone, :qq, :t_shirt)`
	updUserEnableSQL = "UPDATE user SET is_enable=:is_enable WHERE username=:username"

	addTeamSQL        = "INSERT INTO team(name, is_enable, is_self) VALUES(:name, :is_enable, :is_self)"
	addTeamUserRelSQL = "INSERT INTO team_user_rel(team_id, username) VALUES(:team_id, :username)"
	updTeamEnableSQL  = "UPDATE team SET is_enable=:is_enable WHERE id=:id"

	addContestProblemSQL  = "INSERT INTO contest_problem(contest_id, oj_id, pid, `index`) VALUES(:contest_id, :oj_id, :pid, :index)"
	addContestGroupRelSQL = "INSERT INTO contest_group_rel(group_id, contest_id) VALUES(:group_id, :contest_id)"

	getAwardsSQL = `SELECT user.username AS username, medal, award, xcpc_id
FROM user, team_user_rel, xcpc_team_rel, xcpc
WHERE user.username=team_user_rel.username
AND team_user_rel.team_id=xcpc_team_rel.team_id
AND xcpc.id = xcpc_team_rel.xcpc_id`
)
