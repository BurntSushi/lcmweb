function jpost(url, data) {
    return $.post(url, data, function() {}, "json");
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
    $success('timeout', timeout);
}

function flash_error(html) {
    $error = $('#flash_error');
    if ($error.is(':visible')) {
        return;
    }

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

function form_set_error(message) {
    $err = $('#form_error');
    $err.html(message);
    $err.show();
}

function form_response_error(json_response) {
    if (!is_success(json_response)) {
        $err = $('#form_error');
        if (json_response.status && json_response.message) {
            $err.html(json_response.message);
        } else {
            $err.html('Unknown error.');
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
        $('#flash_error').fadeOut();
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
