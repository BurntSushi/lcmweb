$(document).ready(function() {
    var $form = $("#new-password");
    var $submit = $form.find('input[type=submit]');
    var $resend = $form.find('.resend');
    var $password = $form.find('input[name=Password]');
    var $upload = $form.find('#Upload');
    var userid = $form.find('input[name=UserId]').val();

    jajaxForm($form, function (r, status, xhr, $form) {
        if (!is_success(r)) {
            form_response_error(r);
            return;
        }
        form_hide_error();
        console.log(r);
        window.location.replace('/');
    });

    $resend.find('a').click(function() {
        var $stat = $resend.find('.resend-status');

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
