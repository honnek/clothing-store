<?php

namespace App\EventSubscriber;

use App\Event\UserLoggedInViaSocialNetworkEvent;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;

/**
 * @TODO не хватает объекта sender который вызывался бы здесь и отправлял письмо с новым паролем.
 * Должен быть в Utils/Mailer/Sender
 * Также не хватает еще одного шаблона для письма в main/email/client
 */
class UserLoggedInViaSocialNetworkSendNotificationSubscriber implements EventSubscriberInterface
{

    public function __construct() {

    }

    public function onUserLoggedInViaSocialNetworkEvent(UserLoggedInViaSocialNetworkEvent $event)
    {
        $user = $event->getUser();
        $plainPassword = $event->getPlainPassword();
    }

    /**
     * @return string[]
     */
    public static function getSubscribedEvents()
    {
        return [
            UserLoggedInViaSocialNetworkEvent::class => 'onUserLoggedInViaSocialNetworkEvent'
        ];
    }
}
