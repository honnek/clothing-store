<?php

namespace App\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Exception\RuntimeException;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputDefinition;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Input\InputOption;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Question\Question;
use Symfony\Component\Console\Style\SymfonyStyle;
use Symfony\Component\PasswordHasher\Hasher\UserPasswordHasherInterface;
use Symfony\Component\Stopwatch\Stopwatch;
use function Webmozart\Assert\Tests\StaticAnalysis\null;

#[AsCommand(
    name: 'app:add-user',
    description: 'Create user',
)]
class AddUserCommand extends Command
{
    private EntityManagerInterface $entityManager;
    private UserPasswordHasherInterface $userPasswordHasher;
    private UserRepository $userRepository;

    /**
     * @param string|null $name
     */
    public function __construct(
        EntityManagerInterface      $entityManager,
        UserPasswordHasherInterface $userPasswordHasher,
        UserRepository              $userRepository,
        string                      $name = null,
    )
    {
        $this->entityManager = $entityManager;
        $this->userPasswordHasher = $userPasswordHasher;
        $this->userRepository = $userRepository;
        parent::__construct($name);
    }

    protected function configure(): void
    {
        $this
            ->addOption('email', 'em', InputArgument::OPTIONAL, 'Email')
            ->addOption('password', 'p', InputArgument::OPTIONAL, 'Password')
            ->addOption('isAdmin', '', InputArgument::OPTIONAL, 'Admin yes/no', 0);
    }

    /**
     * @param InputInterface $input
     * @param OutputInterface $output
     * @return int
     */
    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $helper = $this->getHelper('question');
        $questionEmail = new Question('Please Enter email: ');
        $email = $helper->ask($input, $output, $questionEmail);

        $questionPassword = new Question('Please Enter password: ');
        $password = $helper->ask($input, $output, $questionPassword);

        $questionIsAdmin = new Question('Please Enter is admin: ', 0);
        $isAdmin = $helper->ask($input, $output, $questionIsAdmin);

        $io = new SymfonyStyle($input, $output);
        $stopwatch = new Stopwatch();
        $stopwatch->start('add-user-command');

        $io->title('Add User Command Wizard');
        $io->text(' Please enter some information');

        $isAdmin = boolval($isAdmin);

        try {
            $newUser = $this->createUser($email, $password, $isAdmin);

        } catch (RuntimeException $exception) {
            $io->comment($exception->getMessage());

            return Command::FAILURE;
        }

        $io->success(sprintf(
            '%s was successful created %s',
            $isAdmin ? 'Administrator user' : 'User',
            $email
        ));

        $event = $stopwatch->stop('add-user-command');
        $stopwatchMessage = sprintf(
            'New user\'s id: %s / Elapsed time: %.2f ms / Consumed memory: %.2f MB',
            $newUser->getId(),
            $event->getDuration(),
            $event->getMemory() / 1000 / 1000
        );
        $io->comment($stopwatchMessage);

        return Command::SUCCESS;
    }

    /**
     * Если пользователя по таким параметрам не существует, то создает его, иначе Exception
     *
     * @param string $email
     * @param string $password
     * @param bool $isAdmin
     * @return User
     */
    private function createUser(string $email, string $password, bool $isAdmin): User
    {
        $existingUser = $this->userRepository->findBy(['email' => $email]);

        if ($existingUser) {
            throw new RuntimeException(message: 'User already existing');
        }

        $user = new User();
        $user->setEmail($email);
        $user->setRoles([$isAdmin ? 'ROLE_ADMIN' : 'ROLE_USER']);
        $user->setPassword($this->userPasswordHasher->hashPassword($user, $password));
        $user->setIsVerified(true);

        $this->entityManager->persist($user);
        $this->entityManager->flush();

        return $user;
    }
}
