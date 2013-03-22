// Simple sprintf taken from: http://goo.gl/iLCwd
// If we need to get fancy, then:
// http://www.diveintojavascript.com/projects/javascript-sprintf
if (!String.prototype.format) {
    String.prototype.format = function() {
        var args = arguments;
        return this.replace(/{(\d+)}/g, function(match, number) { 
            if (typeof args[number] != 'undefined') {
                return args[number];
            }
            return match;
        });
    };
}

// set_viewable_offset takes a DOM element and an event with mouse information,
// and sets the offset (using jQuery) so that the element is viewable within 
// the window.
// The given mouse coordinates are used as the base of where the element should
// be positioned.
function viewable_show($obj, ev, complete) {
    // Give the box some room.
    var spacing = 10;

    var wbot = $(window).scrollTop() + $(window).height();
    var wrht = $(window).scrollLeft() + $(window).width();

    $obj.show();
    var height = $obj.outerHeight();
    var width = $obj.outerWidth();
    $obj.hide();

    var otop = ev.pageY;
    var olft = ev.pageX;
    var obot = otop + height;
    var orht = olft + width;
    $obj.data('origin', 'top');

    if (obot + spacing > wbot) {
        $obj.data('origin', 'bottom');
        otop -= height;
    }
    if (orht + spacing > wrht) {
        olft -= width;
    }

    $obj.show();
    $obj.offset({top: otop, left: olft});
    $obj.hide();

    if ($obj.data('origin') == 'top') {
        $obj.show('slide', {
            complete: complete,
            duration: 200,
            direction: 'up'
        });
    } else {
        $obj.show('slide', {
            complete: complete,
            duration: 200,
            direction: 'down'
        });
    }
}

function viewable_hide($obj, complete) {
    if ($obj.data('origin') == 'top') {
        $obj.hide('slide', {
            complete: complete,
            duration: 200,
            direction: 'up'
        });
    } else {
        $obj.hide('slide', {
            complete: complete,
            duration: 200,
            direction: 'down'
        });
    }
}

function jpost(url, data) {
    return $.post(url, data, function() {}, "json");
}

function html_get(url, data) {
    return $.get(url, data, function() {}, "html").fail(
        function(xhr, text_status, error_thrown) {
            flash_error('<p>Could not load HTML from server:</p>\n' +
                        '<p>' + error_thrown + '</p>');
        }
    );
}

function is_success(json_response) {
    return json_response.status
           && json_response.status == 'success';
}

function flash_success(html) {
    $success = $('#flash_success');
    if ($success.is(':visible')) {
        return;
    }

    $success.find('.flash_message').html(html);
    $success.fadeIn();
    timeout = window.setTimeout(function() {
        $success.fadeOut();
    }, 5 * 1000);
    $success.data('timeout', timeout);
}

function flash_error(html) {
    $error = $('#flash_error');
    $msg = $error.find('.flash_message');
    $msg.html(html);
    $error.fadeIn();
}

function flash_response_error(jr) {
    html = '';
    if (jr.status == 'noauth') {
        html = "<p>Your account is no longer authenticated. Please " +
               "<a href=\"javascript:window.location.replace(window.location);\">" +
               "try refreshing the page</a> and logging back in.</p>";
        if (jr.message && jr.message.length > 0) {
            html += "<p>Could not authenticate because: " + jr.message + "</p>";
        }
    } else if (jr.status == 'fail') {
        html = "<p>" + jr.message + "</p>";
    } else if (jr.status == 'error') {
        html = "<p>An unexpected error has occurred.</p>";
        html += "<p>" + jr.message + "</p>";
    } else if (!is_success(jr)) {
        html = "<p>An unknown error occurred.</p>";
    }
    if (html.length > 0) {
        flash_error(html);
    }
}

function flash_hide_error() {
    $('#flash_error').fadeOut();
}

function form_set_error(message) {
    $err = $('#form_error');
    $err.find('.form_error_message').html('<p>' + message + '</p>');
    $err.show();
}

function form_response_error(json_response) {
    if (!is_success(json_response)) {
        $err = $('#form_error');
        $msg = $err.find('.form_error_message');
        if (json_response.status && json_response.message) {
            $msg.html('<p>' + json_response.message + '</p>');
        } else {
            $msg.html('<p>Unknown error.</p>');
        }

        $err.show();
        return
    }
}

function form_hide_error() {
    $err = $('#form_error');
    $err.hide();
}

function jajaxForm($form, success) {
    $form.ajaxForm({
        success: success,
        dataType: 'json',
        error: function(r, xhr, message, stat) {
            form_set_error('Unknown error (bug): ' + message);
        }
    });
}

function jajaxSubmit($form, data, success) {
    $form.ajaxSubmit({
        success: success,
        dataType: 'json',
        data: data,
        error: function(r, xhr, message, stat) {
            form_set_error('Unknown error (bug): ' + message);
        }
    });
}

$(document).ready(function() {
    $('#flash_error a.flash_dismiss').click(function() {
        flash_hide_error();
    });
    $('#flash_success a.flash_dismiss').click(function() {
        $success = $('#flash_success');
        timeout = $success.data('timeout');
        if (timeout) {
            window.clearTimeout(timeout);
        }
        $success.fadeOut();
    });
});
