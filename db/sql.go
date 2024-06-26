package db

const (
	addUserSQL = `INSERT INTO user(username, nickname, is_admin, is_enable, id_card, phone, qq, t_shirt)
 VALUES(:username, :nickname, :is_admin, :is_enable, :id_card, :phone, :qq, :t_shirt)`
	updUserEnableSQL         = "UPDATE user SET is_enable=:is_enable WHERE username=:username"
	updContestGroupEnableSQL = "UPDATE contest_group SET is_enable=:is_enable WHERE id=:id"
	addContestGroupSQL       = "INSERT INTO contest_group(id, name,is_enable) values(:id, :name, :is_enable)"

	addTeamSQL        = "INSERT INTO team(name, is_enable, is_self) VALUES(:name, :is_enable, :is_self)"
	addTeamUserRelSQL = "INSERT INTO team_user_rel(team_id, username) VALUES(:team_id, :username)"
	updTeamEnableSQL  = "UPDATE team SET is_enable=:is_enable WHERE id=:id"

	addTeamGroupSQL    = "INSERT INTO team_group(group_id, group_name, is_grade) VALUES(:group_id, :group_name, :is_grade)"
	addTeamGroupRelSQL = "INSERT INTO team_group_rel(group_id, team_id) VALUES(:group_id, :team_id)"

	addEventSQL       = "INSERT INTO event(name,start_time,end_time) VALUES(:name, :start_time, :end_time)"
	addHistorySQL     = "INSERT INTO history(name,start_time,end_time,md) VALUES(:name, :start_time, :end_time, :md)"
	addXcpcSQL        = "INSERT INTO xcpc(name,date) VALUES(:name, :date)"
	addXcpcTeamRelSQL = "INSERT INTO xcpc_team_rel(xcpc_id, team_id, medal, award) VALUES(:xcpc_id, :team_id, :medal, :award)"

	addContestProblemSQL  = "INSERT INTO contest_problem(contest_id, oj_id, pid, `index`) VALUES(:contest_id, :oj_id, :pid, :index)"
	addContestGroupRelSQL = "INSERT INTO contest_group_rel(group_id, contest_id) VALUES(:group_id, :contest_id)"
	addContestTeamRelSQL  = "INSERT INTO contest_team_rel(contest_id, team_id) VALUES(:contest_id, :team_id)"

	getAwardsSQL = `SELECT user.username AS username, medal, award, xcpc_id
FROM user, team_user_rel, xcpc_team_rel, xcpc
WHERE user.username=team_user_rel.username
AND team_user_rel.team_id=xcpc_team_rel.team_id
AND xcpc.id = xcpc_team_rel.xcpc_id`
)
