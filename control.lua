 -- This file stands for console logging for FactoCord-3.0 integration
 -- Please configure as needed, any discord message will be sent in
 --  raw format if it starts with `0000-00-00 00:00:00 [DISCORD] `
 -- For more information visit https://github.com/maxsupermanhd/FactoCord-3.0
 -- If you have any question or comments join our Discord https://discord.gg/uNhtRH8

local FactoCordIntegration = {}

function FactoCordIntegration.PrintToDiscord(msg)
	localised_print({"", "0000-00-00 00:00:00 [DISCORD] ", msg})
end

-- Send a private message to a specific player (not visible in logs when used correctly)
-- This uses player.print() which only shows to that specific player
function FactoCordIntegration.WhisperToPlayer(player_name, message, color)
	local p = game.get_player(player_name)
	if p and p.connected then
		local print_color = color or {r=0.5, g=0.8, b=1.0} -- Light blue default
		p.print(message, {color=print_color})
		return true
	end
	return false
end

-- Send verification code to player (private, not logged)
-- This should be called via /silent-command to avoid logging
function FactoCordIntegration.SendVerificationCode(player_name, code)
	local p = game.get_player(player_name)
	if p and p.connected then
		-- Using player.print() ensures only this player sees the message
		p.print("[FactoCord Verification]", {color={r=0.2,g=0.9,b=0.2}})
		p.print("Your verification code: " .. code, {color={r=1,g=1,b=0}})
		p.print("Enter this code in Discord DM to verify your account.", {color={r=0.7,g=0.7,b=0.7}})
		return true
	end
	return false
end

-- Execute a command silently as a specific player
-- This uses game.players[x].print for responses and avoids global prints
function FactoCordIntegration.SilentExecute(player_name, lua_code)
	local p = game.get_player(player_name)
	if not p then
		return false, "Player not found"
	end
	
	-- Execute the Lua code in a protected call
	local func, err = load(lua_code)
	if not func then
		return false, "Syntax error: " .. tostring(err)
	end
	
	local success, result = pcall(func)
	if not success then
		return false, "Execution error: " .. tostring(result)
	end
	
	return true, result
end

-- Check if a player has spy mode enabled for a specific log type
-- This function can be used by logging mods to check if they should suppress logs
-- Usage: if FactoCordIntegration.IsSpyModeActive(player_name, "BUILT_ENTITY") then return end
function FactoCordIntegration.IsSpyModeActive(player_name, log_type)
	if not storage.factocord_spy_mode then
		return false
	end
	
	local player_spy = storage.factocord_spy_mode[player_name]
	if not player_spy then
		return false
	end
	
	-- Check if "ALL" is in the list or the specific log type
	for _, suppressed_type in pairs(player_spy) do
		if suppressed_type == "ALL" or suppressed_type == log_type then
			return true
		end
	end
	
	return false
end

-- Get all players with spy mode active
function FactoCordIntegration.GetSpyModePlayers()
	if not storage.factocord_spy_mode then
		return {}
	end
	return storage.factocord_spy_mode
end

-- Set spy mode for a player (called from Discord bot via silent-command)
function FactoCordIntegration.SetSpyMode(player_name, log_types)
	if not storage.factocord_spy_mode then
		storage.factocord_spy_mode = {}
	end
	
	if log_types == nil or #log_types == 0 then
		storage.factocord_spy_mode[player_name] = nil
	else
		storage.factocord_spy_mode[player_name] = log_types
	end
end

script.on_event(defines.events.on_player_joined_game, function(event)
	local p = game.players[event.player_index];
	FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** joined.");
	if(p.admin == true) then
		p.print("Welcome admin " .. p.name .. " to server!");
	else
		p.print("Welcome " .. p.name .. " to server!");
	end
end)

script.on_event(defines.events.on_player_left_game, function(event)
	local p = game.players[event.player_index];
	FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** left.");
end)

script.on_event({defines.events.on_console_chat},
	function (e)
		if not e.player_index then
			return
		end
		if game.players[e.player_index].tag == "" then
			if game.players[e.player_index].admin then
				FactoCordIntegration.PrintToDiscord('(Admin) <' .. game.players[e.player_index].name .. '> ' .. e.message)
			else
				FactoCordIntegration.PrintToDiscord('<' .. game.players[e.player_index].name .. '> ' .. e.message)
			end
		else
			if game.players[e.player_index].admin then
				FactoCordIntegration.PrintToDiscord('(Admin) <' .. game.players[e.player_index].name .. '> ' .. game.players[e.player_index].tag .. " " .. e.message)
			else
				FactoCordIntegration.PrintToDiscord('<' .. game.players[e.player_index].name .. '> ' .. game.players[e.player_index].tag .. " " .. e.message)
			end
		end
	end
)


script.on_event(defines.events.on_player_died, function(event)
	local p = game.players[event.player_index];
	local c = event.cause
	if not c then
		FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** died.");
	else
		local name = "Unknown";
		if c.type == "character" then
			name = c.player.name;
		elseif c.type == "spider-vehicle" then
			if c.entity_label then
				name = {"", c.localised_name, " " , c.entity_label};
			else
				name = {"", "a ", c.localised_name};
			end
		elseif c.type == "locomotive" then
			name = {"", c.localised_name, " " , c.backer_name};
		else
			name = {"", "a ", c.localised_name};
		end
		FactoCordIntegration.PrintToDiscord({"", "**", p.name, "** was killed by ", name, "."});
	end
end)
script.on_event(defines.events.on_player_kicked, function(event)
	local p = game.players[event.player_index];
	FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** kicked.");
end)
script.on_event(defines.events.on_player_unbanned, function(event)
	FactoCordIntegration.PrintToDiscord("**" .. event.player_name .. "** unbanned.");
end)
script.on_event(defines.events.on_player_unmuted, function(event)
	local p = game.players[event.player_index];
	FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** unmuted.");
end)
script.on_event(defines.events.on_player_banned, function(event)
	FactoCordIntegration.PrintToDiscord("**" .. event.player_name .. "** banned.");
end)
script.on_event(defines.events.on_player_muted, function(event)
	local p = game.players[event.player_index];
	FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** muted.");
end)


return FactoCordIntegration;
