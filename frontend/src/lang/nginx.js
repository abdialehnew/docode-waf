import { StreamLanguage } from '@codemirror/language';

// Simple Nginx config language mode
export const nginx = StreamLanguage.define({
  startState: () => ({
    inComment: false,
    inString: false,
  }),
  
  token: (stream, state) => {
    // Comments
    if (stream.match(/^#.*/)) {
      return 'comment';
    }
    
    // Strings
    if (stream.match(/^"(?:[^"\\]|\\.)*"/)) {
      return 'string';
    }
    if (stream.match(/^'(?:[^'\\]|\\.)*'/)) {
      return 'string';
    }
    
    // Directives (keywords that end with semicolon or brace)
    if (stream.match(/^(server|location|upstream|http|events|stream|mail|types|if|return|rewrite|set|proxy_pass|proxy_set_header|root|index|listen|server_name|ssl_certificate|ssl_certificate_key|include|error_page|access_log|error_log|gzip|gzip_types|client_max_body_size|proxy_buffer_size|proxy_buffers|proxy_connect_timeout|proxy_read_timeout|proxy_send_timeout|keepalive_timeout|send_timeout|fastcgi_pass|uwsgi_pass|add_header|limit_req|limit_req_zone|limit_conn|limit_conn_zone|allow|deny|auth_basic|auth_basic_user_file|ssl_protocols|ssl_ciphers|ssl_prefer_server_ciphers|ssl_session_cache|ssl_session_timeout|ssl_stapling|ssl_stapling_verify|resolver|try_files|autoindex|expires|charset|default_type)\b/)) {
      return 'keyword';
    }
    
    // Blocks
    if (stream.match(/^[{}]/)) {
      return 'brace';
    }
    
    // Variables
    if (stream.match(/^\$[a-zA-Z_][a-zA-Z0-9_]*/)) {
      return 'variableName';
    }
    
    // Numbers
    if (stream.match(/^\d+[KkMmGg]?/)) {
      return 'number';
    }
    
    // Operators
    if (stream.match(/^[~=]/)) {
      return 'operator';
    }
    
    // Paths and URLs
    if (stream.match(/^\/[^\s;{]*/)) {
      return 'url';
    }
    if (stream.match(/^https?:\/\/[^\s;{]*/)) {
      return 'url';
    }
    
    // Boolean values
    if (stream.match(/^(on|off)\b/)) {
      return 'bool';
    }
    
    stream.next();
    return null;
  },
});
