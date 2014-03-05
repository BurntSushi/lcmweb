$(document).ready(function() {
    function ping() {
        jpost('/noop').always(function(r) {
            if (!is_success(r)) {
                flash_response_error(r);
            } else {
                flash_hide_error();
            }
        });
    }
    window.setInterval(ping, 30 * 1000);
});
