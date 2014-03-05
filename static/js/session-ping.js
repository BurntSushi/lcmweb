$(document).ready(function() {
    function ping() {
        jpost('/noop').always(function(r) {
            if (!is_success(r)) {
                flash_response_error(r);
            }
            // Don't hide the error if successful since there may be
            // an unrelated error already on the screen.
        });
    }
    window.setInterval(ping, 30 * 1000);
});
