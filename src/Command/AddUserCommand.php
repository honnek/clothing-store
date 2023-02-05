<?php

namespace App\Command;

use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Input\InputOption;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Question\Question;
use Symfony\Component\Console\Style\SymfonyStyle;
use Symfony\Component\Stopwatch\Stopwatch;

#[AsCommand(
    name: 'app:add-user',
    description: 'Create user',
)]
class AddUserCommand extends Command
{
    protected function configure(): void
    {
        $this
            ->addOption('email', 'em', InputArgument::OPTIONAL, 'Email')
            ->addOption('password', 'p', InputArgument::OPTIONAL, 'Password')
            ->addOption('isAdmin', '', InputArgument::OPTIONAL, 'Admin yes/no', 0);
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $io = new SymfonyStyle($input, $output);
        $stopwatch = new Stopwatch();
        $stopwatch->start('add-user-command');
        $email = $input->getOption('email');
        $password = $input->getOption('password');
        $isAdmin = $input->getOption('isAdmin');

        $io->title('Add User Command Wizard');
        $io->text(' Please enter some information');

        if (!$email) {
            $io->ask('Enter email');
        }

        if (!$password) {
            $io->askHidden('Enter password');
        }

        if (!$isAdmin) {
            $io->askQuestion(new Question('User is admin? (1 or 0)', 0));
        }


//dd();
//        if ($arg1) {
//            $io->note(sprintf('You passed an argument: %s', $arg1));
//        }
//
//        if ($input->getOption('option1')) {
//            // ...
//        }
//
//        $io->success('You have a new command! Now make it your own! Pass --help to see your options.');
        $io->success(sprintf(
            '%s was successful created %s',
            $isAdmin ? 'Administrator user' : 'User',
            $email
        ));

        $event = $stopwatch->stop('add-user-command');
        $stopwatchMessage = sprintf(
            'New user\'s id: %s / Elapsed time: %.2f ms / Consumed memory: %.2f MB',
            12,
            $event->getDuration(),
            $event->getMemory()/1000/1000
        );
        $io->comment($stopwatchMessage);

        return Command::SUCCESS;
    }
}
