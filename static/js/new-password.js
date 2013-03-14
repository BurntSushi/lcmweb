function jpost(url, data) {
    return $.post(url, data, function() {}, "json");
}

function is_success(json_response) {
    return json_response.status
           && json_response.status == 'success';
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
    $form = $("#new-password");
    $submit = $form.find('input[type=submit]');
    $resend = $form.find('.resend');
    $password = $form.find('input[name=Password]');
    $upload = $form.find('#Upload');
    userid = $form.find('input[name=UserId]').val();

    jajaxForm($form, function (r, status, xhr, $form) {
        if (!is_success(r)) {
            form_response_error(r);
            return;
        }
        form_hide_error();
        console.log(r);
    });

    $resend.find('a').click(function() {
        $stat = $resend.find('.resend-status');

        $stat.removeClass('success');
        $stat.addClass('pending');
        $stat.text('Sending...');
        $stat.show();

        jpost("/newpassword-send", { "UserId": userid })
        .done(function(r) {
            if (!is_success(r)) {
                $stat.hide();
                form_response_error(r);
                return;
            }

            $stat.fadeOut({
                done: function() {
                    $stat.removeClass('pending');
                    $stat.addClass('success');
                    $stat.text('Success!');
                    $stat.fadeIn();
                }
            });
        });
    });
});
