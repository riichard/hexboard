-- hexboard.lua — Neovim cursor integration
--
-- Mirrors your editor cursor position to the hexboard display in real time.
-- The display has 4 rows × 32 columns; editor line/col are mapped with modulo.
--
-- Usage (in init.lua):
--   require('path.to.hexboard').setup({ host = 'txt', port = 8082 })
--
-- Or source directly:
--   vim.cmd('source /path/to/hexboard.lua')
--   require('hexboard').setup()

local M = {}

local uv = vim.uv or vim.loop

local config = {
	host = 'txt',
	port = 8082,
}

local conn    = nil  -- active TCP handle
local pending = false

local function close_conn()
	if conn then
		pcall(function() conn:close() end)
		conn = nil
	end
end

local function connect()
	close_conn()
	local sock = uv.new_tcp()
	sock:connect(config.host, config.port, function(err)
		if err then
			pcall(function() sock:close() end)
			return
		end
		conn = sock
	end)
end

local function send_position()
	pending = false
	if not conn then
		connect()
		return
	end
	-- Map editor position onto the 32×4 display grid.
	-- row: editor line (1-based) → 0-3 via modulo
	-- col: editor column (0-based) → 0-31 via modulo
	local pos = vim.api.nvim_win_get_cursor(0)
	local col = pos[2] % 32
	local row = (pos[1] - 1) % 4
	conn:write(col .. ' ' .. row .. '\n', function(err)
		if err then close_conn() end
	end)
end

-- Debounce cursor events: only send after 40 ms of inactivity
-- to avoid saturating the TCP stream during fast movement.
local function on_cursor_moved()
	if not pending then
		pending = true
		vim.defer_fn(function()
			if vim.api.nvim_get_mode().mode ~= 'c' then
				send_position()
			else
				pending = false
			end
		end, 40)
	end
end

function M.setup(opts)
	if opts then
		config = vim.tbl_extend('force', config, opts)
	end
	connect()
	vim.api.nvim_create_autocmd({ 'CursorMoved', 'CursorMovedI' }, {
		desc  = 'hexboard: sync cursor position',
		callback = on_cursor_moved,
	})
	-- Reconnect if the connection drops
	vim.api.nvim_create_autocmd('FocusGained', {
		desc = 'hexboard: reconnect on focus',
		callback = function()
			if not conn then connect() end
		end,
	})
end

return M
