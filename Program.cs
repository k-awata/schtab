using System.CommandLine;
using System.Diagnostics;
using Microsoft.Win32.TaskScheduler;

namespace SchtabCommand;

class Program
{
    const string NewSchtab = @"#  +---------------- Minute            (range: 0-59)
#  |  +------------- Hour              (range: 0-23)
#  |  |  +---------- Day of the Month  (range: 1-31)
#  |  |  |  +------- Month of the Year (range: 1-12)
#  |  |  |  |  +---- Day of the Week   (range: 1-7, 1 standing for Monday)
#  |  |  |  |  |
#  *  *  *  *  *  cmd [args...]
";

    static async Task<int> Main(string[] args)
    {
        var rootCommand = new RootCommand("schtab sets tasks to Windows Task Scheduler from a text in crontab format.");
        var fileArgument = new Argument<FileInfo?>("file", () => null, "File in crontab format, when file is -, read standard input");
        var editOption = new Option<bool>("-e", "Edit your schtab");
        var listOption = new Option<bool>("-l", "List your schtab");
        var remOption = new Option<bool>("--delete", "Delete your schtab");
        var regOption = new Option<bool>("--reg", "Register tasks in your schtab at Task Scheduler");
        var unregOption = new Option<bool>("--unreg", "Unregister tasks from Task Scheduler");
        var schtabOption = new Option<FileInfo>("--schtab", "Path of your schtab file");
        schtabOption.SetDefaultValue(new FileInfo(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData) + @"\schtab"));

        rootCommand.AddArgument(fileArgument);
        rootCommand.AddOption(editOption);
        rootCommand.AddOption(listOption);
        rootCommand.AddOption(regOption);
        rootCommand.AddOption(unregOption);
        rootCommand.AddOption(remOption);
        rootCommand.AddOption(schtabOption);

        rootCommand.SetHandler((file, edit, list, rem, reg, unreg, schtab) =>
        {
            if (list)
            {
                if (!schtab.Exists)
                {
                    Console.Error.WriteLine("No schtab file");
                    return;
                }
                Console.Write(File.ReadAllText(schtab.FullName));
                return;
            }

            if (edit)
            {
                if (!schtab.Exists)
                {
                    File.WriteAllText(schtab.FullName, NewSchtab);
                }
                try
                {
                    using (var proc = new Process())
                    {
                        var editor = (
                            Environment.GetEnvironmentVariable("VISUAL") ??
                            Environment.GetEnvironmentVariable("EDITOR") ??
                            "notepad").Split(new char[0], StringSplitOptions.RemoveEmptyEntries);
                        proc.StartInfo.FileName = editor[0];
                        foreach (var arg in editor[1..])
                        {
                            proc.StartInfo.ArgumentList.Add(arg);
                        }
                        proc.StartInfo.ArgumentList.Add(schtab.FullName);
                        proc.StartInfo.UseShellExecute = true;
                        proc.StartInfo.CreateNoWindow = true;
                        proc.Start();
                        proc.WaitForExit();
                    }
                    registerTasks(schtab.FullName);
                }
                catch (Exception e)
                {
                    Console.Error.WriteLine(e);
                }
                return;
            }

            if (reg)
            {
                if (!schtab.Exists)
                {
                    Console.Error.WriteLine("No schtab file");
                    return;
                }
                registerTasks(schtab.FullName);
                return;
            }

            if (unreg)
            {
                unregisterTasks();
                return;
            }

            if (rem)
            {
                if (!schtab.Exists)
                {
                    Console.Error.WriteLine("No schtab file");
                    return;
                }
                schtab.Delete();
                unregisterTasks();
                return;
            }

            if (file == null)
            {
                Console.Error.WriteLine("No input file");
                return;
            }

            if (file.Name == "-")
            {
                File.WriteAllText(schtab.FullName, Console.In.ReadToEnd());
            }
            else if (file.Exists)
            {
                File.WriteAllText(schtab.FullName, File.ReadAllText(file.FullName));
            }
            else
            {
                Console.Error.WriteLine($"Could not find file \"{file.Name}\"");
                return;
            }

            registerTasks(schtab.FullName);
        }, fileArgument, editOption, listOption, remOption, regOption, unregOption, schtabOption);

        return await rootCommand.InvokeAsync(args);
    }

    internal static void registerTasks(string file)
    {
        try
        {
            Task.UnregisterAll();
            var tasks = new List<Task>();
            var tn = "";
            foreach (var l in File.ReadLines(file))
            {
                var trimmed = l.Trim();
                if (trimmed.StartsWith('#'))
                {
                    tn = trimmed.Substring(1).Trim();
                }
                else if (trimmed != "")
                {
                    if (tn == "" || tn.IndexOfAny(Path.GetInvalidFileNameChars()) != -1)
                    {
                        tn = "Task" + (tasks.Count + 1).ToString("D3");
                    }
                    tasks.Add(new Task(tn, trimmed));
                    tn = "";
                }
            }
            foreach (var t in tasks)
            {
                t.Register();
            }
            Console.WriteLine("schtab registered your tasks at Task Scheduler");
        }
        catch (Exception e)
        {
            Console.Error.WriteLine(e);
        }
    }

    internal static void unregisterTasks()
    {
        try
        {
            Task.UnregisterAll();
            Console.WriteLine("schtab unregistered all your tasks from Task Scheduler");
        }
        catch (Exception e)
        {
            Console.Error.WriteLine(e);
        }
    }
}

class Task
{
    private static readonly string rootPath = "schtab";
    private static readonly string userPath = Environment.UserName + "@" + Environment.UserDomainName;

    public static void UnregisterAll()
    {
        try
        {
            var tf = TaskService.Instance.RootFolder.SubFolders[rootPath].SubFolders[userPath];
            foreach (var t in tf.AllTasks)
            {
                tf.DeleteTask(t.Name);
            }
        }
        catch { }
    }

    private readonly string path;
    private readonly Trigger trigger;
    private readonly ExecAction action;

    public Task(string name, string cron)
    {
        var args = cron.Split(new char[0], 7, StringSplitOptions.RemoveEmptyEntries);
        path = rootPath + @"\" + userPath + @"\" + name;
        trigger = Trigger.FromCronFormat(String.Join(' ', args, 0, 5))[0];
        if (args.Length == 6)
        {
            action = new ExecAction(args[5]);
        }
        else
        {
            action = new ExecAction(args[5], args[6]);
        }
    }

    public void Register()
    {
        TaskService.Instance.AddTask(path, trigger, action);
    }
}