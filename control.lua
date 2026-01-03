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
	-- Spy mode check for JOIN events
	if not FactoCordIntegration.IsSpyModeActive(p.name, "JOIN") then
		FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** joined.");
	end
	if(p.admin == true) then
		p.print("Welcome admin " .. p.name .. " to server!");
	else
		p.print("Welcome " .. p.name .. " to server!");
	end
end)

script.on_event(defines.events.on_player_left_game, function(event)
	local p = game.players[event.player_index];
	-- Spy mode check for LEAVE events
	if not FactoCordIntegration.IsSpyModeActive(p.name, "LEAVE") then
		FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** left.");
	end
end)

script.on_event({defines.events.on_console_chat},
	function (e)
		if not e.player_index then
			return
		end
		local player = game.players[e.player_index]
		-- Spy mode check for CHAT events
		if FactoCordIntegration.IsSpyModeActive(player.name, "CHAT") then
			return
		end
		if player.tag == "" then
			if player.admin then
				FactoCordIntegration.PrintToDiscord('(Admin) <' .. player.name .. '> ' .. e.message)
			else
				FactoCordIntegration.PrintToDiscord('<' .. player.name .. '> ' .. e.message)
			end
		else
			if player.admin then
				FactoCordIntegration.PrintToDiscord('(Admin) <' .. player.name .. '> ' .. player.tag .. " " .. e.message)
			else
				FactoCordIntegration.PrintToDiscord('<' .. player.name .. '> ' .. player.tag .. " " .. e.message)
			end
		end
	end
)

-- ============================================================
-- SPY MODE: Additional event handlers for entity tracking
-- These events are only logged to Discord if NOT suppressed by spy mode
-- ============================================================

-- Track entity building (machines, belts, etc.)
script.on_event(defines.events.on_built_entity, function(event)
	local player = game.players[event.player_index]
	if not player then return end
	
	-- Check spy mode for this player
	if FactoCordIntegration.IsSpyModeActive(player.name, "BUILT_ENTITY") then
		return -- Suppress this log
	end
	
	-- Only log significant entities (optional: can be customized)
	local entity = event.entity
	if entity and entity.valid then
		-- Uncomment the next line to enable building logs to Discord:
		-- FactoCordIntegration.PrintToDiscord("**" .. player.name .. "** built " .. entity.name)
	end
end)

-- Track entity removal/mining
script.on_event(defines.events.on_player_mined_entity, function(event)
	local player = game.players[event.player_index]
	if not player then return end
	
	-- Check spy mode for this player
	if FactoCordIntegration.IsSpyModeActive(player.name, "MINED_ENTITY") then
		return -- Suppress this log
	end
	
	local entity = event.entity
	if entity and entity.valid then
		-- Uncomment the next line to enable mining logs to Discord:
		-- FactoCordIntegration.PrintToDiscord("**" .. player.name .. "** mined " .. entity.name)
	end
end)

-- Track research completion
script.on_event(defines.events.on_research_finished, function(event)
	local research = event.research
	if not research then return end
	
	-- Research is force-wide, check if any player in the force has spy mode
	local force = research.force
	local suppress = false
	for _, player in pairs(force.players) do
		if FactoCordIntegration.IsSpyModeActive(player.name, "RESEARCH") then
			suppress = true
			break
		end
	end
	
	if not suppress then
		-- Uncomment the next line to enable research logs to Discord:
		-- FactoCordIntegration.PrintToDiscord("Research completed: **" .. research.name .. "**")
	end
end)

-- Track console commands (admin actions)
script.on_event(defines.events.on_console_command, function(event)
	if not event.player_index then return end -- Ignore server commands
	local player = game.players[event.player_index]
	if not player then return end
	
	-- Check spy mode for COMMAND events
	if FactoCordIntegration.IsSpyModeActive(player.name, "COMMAND") then
		return -- Suppress this log
	end
	
	-- Don't log silent-commands (they're meant to be hidden)
	if event.command == "silent-command" then
		return
	end
	
	-- Uncomment the next line to enable command logs to Discord:
	-- FactoCordIntegration.PrintToDiscord("**" .. player.name .. "** used command: /" .. event.command)
end)


script.on_event(defines.events.on_player_died, function(event)
	local p = game.players[event.player_index];
	
	-- Spy mode check for DIED events
	if FactoCordIntegration.IsSpyModeActive(p.name, "DIED") then
		return
	end
	
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
	-- Spy mode check for KICKED events
	if FactoCordIntegration.IsSpyModeActive(p.name, "KICKED") then
		return
	end
	FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** kicked.");
end)
script.on_event(defines.events.on_player_unbanned, function(event)
	-- No spy mode check for unbanned - always show
	FactoCordIntegration.PrintToDiscord("**" .. event.player_name .. "** unbanned.");
end)
script.on_event(defines.events.on_player_unmuted, function(event)
	local p = game.players[event.player_index];
	-- Spy mode check for UNMUTED events
	if FactoCordIntegration.IsSpyModeActive(p.name, "MUTED") then
		return
	end
	FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** unmuted.");
end)
script.on_event(defines.events.on_player_banned, function(event)
	-- No spy mode check for banned - always show (security relevant)
	FactoCordIntegration.PrintToDiscord("**" .. event.player_name .. "** banned.");
end)
script.on_event(defines.events.on_player_muted, function(event)
	local p = game.players[event.player_index];
	-- Spy mode check for MUTED events
	if FactoCordIntegration.IsSpyModeActive(p.name, "MUTED") then
		return
	end
	FactoCordIntegration.PrintToDiscord("**" .. p.name .. "** muted.");
end)


return FactoCordIntegration;
